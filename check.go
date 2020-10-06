package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/segfaultax/go-nagios"
)

const (
	okMsg     = "%d instances of %s found"
	notOkMsg  = "%d instances of %s found, expected at least %d"
	useragent = "otpl-service-check/2.0.0"
)

type check struct {
	cli             *http.Client
	discoveryState  []Announcement
	service         string
	endpoint        string
	skipHealthcheck bool
	warn, crit      int
	headers         []header
}

type response struct {
	endpoint    *url.URL
	contentType string
	body        []byte
	statusCode  int
	duration    time.Duration
}

func (c *check) run(ctx context.Context, acc *resultAccumulator) {
	var matching []Announcement
	for _, a := range c.discoveryState {
		if a.ServiceType == c.service {
			matching = append(matching, a)
		}
	}

	c.checkQuota(matching, acc)

	if !c.skipHealthcheck {
		c.checkInstances(ctx, matching, acc)
	}
}

func (c *check) checkQuota(anns []Announcement, acc *resultAccumulator) {
	seen := make(map[string]bool)
	cnt := 0
	for _, ann := range anns {
		tok := ann.serverToken()
		if ok := seen[tok]; tok == "" || (tok != "" && !ok) {
			seen[tok] = true
			cnt++
		}
	}

	res := result{}
	if c.crit > 0 && cnt < c.crit {
		res.status = nagios.StatusCrit
		res.message = fmt.Sprintf(notOkMsg, cnt, c.service, c.crit)
	} else if c.warn > 0 && cnt < c.warn {
		res.status = nagios.StatusWarn
		res.message = fmt.Sprintf(notOkMsg, cnt, c.service, c.warn)
	} else {
		res.status = nagios.StatusOK
		res.message = fmt.Sprintf(okMsg, cnt, c.service)
	}

	res.perf = append(res.perf, nagios.NewPerfData("instances", float64(cnt), ""))
	acc.add(res)
}

func (c *check) checkInstances(ctx context.Context, anns []Announcement, acc *resultAccumulator) {
	var wg sync.WaitGroup
	for _, ann := range anns {
		wg.Add(1)
		go func(ann Announcement) {
			defer wg.Done()
			c.checkAnnouncement(ctx, ann, acc)
		}(ann)
	}
	wg.Wait()
}

func (c *check) checkAnnouncement(ctx context.Context, ann Announcement, acc *resultAccumulator) {
	resp, err := c.fetchAnnouncement(ctx, ann)
	if err != nil {
		acc.add(result{
			status:  nagios.StatusWarn,
			message: fmt.Sprintf("failed to fetch announced endpoint %s: %s", ann.ServiceURI, err),
		})
		return
	}

	acc.add(result{
		status:  statusFor(resp.statusCode),
		message: formatMessage(resp),
	})
}

func (c *check) fetchAnnouncement(ctx context.Context, ann Announcement) (*response, error) {
	base, err := url.Parse(ann.ServiceURI)
	if err != nil {
		return nil, err
	}

	endpoint, err := base.Parse(c.endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", useragent)
	for _, h := range c.headers {
		req.Header.Add(h.key, h.value)
	}

	start := time.Now()
	resp, err := c.cli.Do(req)
	duration := time.Now().Sub(start)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &response{
		endpoint:    endpoint,
		contentType: resp.Header.Get("Content-Type"),
		body:        body,
		statusCode:  resp.StatusCode,
		duration:    duration,
	}, nil
}

func statusFor(code int) nagios.Status {
	switch code / 100 {
	case 2:
		return nagios.StatusOK
	case 4:
		return nagios.StatusWarn
	default:
		return nagios.StatusCrit
	}
}

func formatMessage(resp *response) string {
	template := `---
status code: %d
duration: %d ms
endpoint: %s
`

	return fmt.Sprintf(template, resp.statusCode, resp.duration.Milliseconds(), resp.endpoint.String())
}
