package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeaders(t *testing.T) {
	assert := assert.New(t)

	hds, err := parseHeaders([]string{"foo: bar", "baz: spam"})
	if err != nil {
		t.Error("unexpected error", err)
	}
	assert.ElementsMatch([]header{
		{key: "foo", value: "bar"},
		{key: "baz", value: "spam"},
	}, hds)

	hds, err = parseHeaders([]string{"invalid", "baz: spam"})
	if err == nil {
		t.Fatal("expected error, got", hds)
	}
}
