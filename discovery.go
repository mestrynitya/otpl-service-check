package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const (
	tokenkey = "server-token"
)

type discoveryClient struct {
	baseURL *url.URL
	client  *http.Client
}

type Announcement struct {
	AnnouncementID string                 `json:"announcementId,omitempty"`
	ServiceType    string                 `json:"serviceType,omitempty"`
	ServiceURI     string                 `json:"serviceUri,omitempty"`
	Environment    string                 `json:"environment,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

func newDiscoveryClient(server string) (*discoveryClient, error) {
	u, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	cli := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &discoveryClient{
		baseURL: u,
		client:  cli,
	}, nil
}

func (d *discoveryClient) findAnnouncements(ctx context.Context) ([]Announcement, error) {
	var anns []Announcement
	err := d.get(ctx, "/state", &anns)
	if err != nil {
		return nil, err
	}
	return anns, nil
}

func (d *discoveryClient) get(ctx context.Context, path string, dst interface{}) error {
	u, err := d.baseURL.Parse(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	return d.do(ctx, req, dst)
}

func (d *discoveryClient) do(ctx context.Context, req *http.Request, dst interface{}) error {
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (a Announcement) serverToken() string {
	if v, ok := a.Metadata[tokenkey]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
