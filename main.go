package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/segfaultax/go-nagios"
	"github.com/spf13/pflag"
)

type header struct {
	key   string
	value string
}

func main() {
	discovery := pflag.StringP("discovery", "d", "", "discovery server URL")
	service := pflag.StringP("service", "s", "", "service name to check")
	endpoint := pflag.StringP("endpoint", "e", "health", "healthcheck endpoint")
	skipHealthcheck := pflag.BoolP("no-healthcheck", "n", false, "disable healthcheck")
	timeout := pflag.IntP("timeout", "t", 5, "http timeout for health endpoint (in seconds)")
	crit := pflag.IntP("crit-fewer", "c", 1, "minimum instances before critical, 0 to disable")
	warn := pflag.IntP("warn-fewer", "w", 1, "minimum instances before warning, 0 to disable")
	headers := pflag.StringSliceP("header", "H", nil, "http headers for health endpoint (eg: 'Accept: application/json')")

	pflag.Parse()

	if *discovery == "" {
		usageErrorAndExit(nagios.StatusUnknown.ExitCode, true, "discovery name is required")
	}

	if *service == "" {
		usageErrorAndExit(nagios.StatusUnknown.ExitCode, true, "service is required")
	}

	if *timeout <= 0 {
		usageErrorAndExit(nagios.StatusUnknown.ExitCode, false, "timeout must be greater than zero")
	}

	if *crit < 0 {
		*crit = 0
	}
	if *warn < 0 {
		*warn = 0
	}
	if *warn < *crit {
		usageErrorAndExit(nagios.StatusUnknown.ExitCode, false, "warn must be greater than crit")
	}

	hds, err := parseHeaders(*headers)
	if err != nil {
		usageErrorAndExit(nagios.StatusUnknown.ExitCode, false, "failed to parse headers: %s", err)
	}

	// start check //

	c := nagios.NewCheck()
	defer c.Done()

	ctx := context.Background()

	disco, err := newDiscoveryClient(*discovery)
	if err != nil {
		c.Unknown("failed to construct discovery client: %s", err)
		return
	}

	anns, err := disco.findAnnouncements(ctx)
	if err != nil {
		c.Unknown("failed to fetch discovery state: %s", err)
		return
	}

	check := &check{
		cli: &http.Client{
			Timeout: time.Duration(*timeout) * time.Second,
		},
		announcements:   anns,
		service:         *service,
		endpoint:        *endpoint,
		skipHealthcheck: *skipHealthcheck,
		warn:            *warn,
		crit:            *crit,
		headers:         hds,
	}

	acc := newAccumulator()
	check.run(ctx, acc)

	acc.updateCheck(c, func(s []string) string {
		return strings.Join(s, "\n")
	})
}

func parseHeaders(hs []string) ([]header, error) {
	var hds []header
	for _, h := range hs {
		if !strings.Contains(h, ":") {
			return nil, fmt.Errorf("invalid header: %s", h)
		}
		parts := strings.SplitN(h, ":", 2)
		hds = append(hds, header{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
	}
	return hds, nil
}

func usageErrorAndExit(code int, showHelp bool, format string, params ...interface{}) {
	msg := fmt.Sprintf(format, params...)
	fmt.Println(msg)
	if showHelp {
		pflag.PrintDefaults()
	}
	os.Exit(code)
}
