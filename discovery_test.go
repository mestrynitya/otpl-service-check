package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoveryClient(t *testing.T) {
	assert := assert.New(t)

	anns := []Announcement{
		{
			AnnouncementID: "ann1",
			ServiceType:    "foo",
			ServiceURI:     "http://foo.com",
			Environment:    "test-rs",
		},
		{
			AnnouncementID: "ann2",
			ServiceType:    "bar",
			ServiceURI:     "http://bar.com",
			Environment:    "test-rs",
		},
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/state" {
			t.Error("expected path to be /state, got", r.URL)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(anns)
	}))

	c, err := newDiscoveryClient(s.URL)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	res, err := c.findAnnouncements(context.TODO())
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	assert.ElementsMatch(anns, res)
}
