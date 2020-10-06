package main

import (
	"strings"
	"testing"

	"github.com/segfaultax/go-nagios"
)

func TestAccumulator(t *testing.T) {
	tests := []struct {
		add      nagios.Status
		expected nagios.Status
	}{
		{
			add:      nagios.StatusUnknown,
			expected: nagios.StatusUnknown,
		},
		{
			add:      nagios.StatusOK,
			expected: nagios.StatusUnknown,
		},
		{
			add:      nagios.StatusWarn,
			expected: nagios.StatusWarn,
		},
		{
			add:      nagios.StatusOK,
			expected: nagios.StatusWarn,
		},
		{
			add:      nagios.StatusCrit,
			expected: nagios.StatusCrit,
		},
	}

	acc := newAccumulator()
	for _, x := range tests {
		acc.add(result{
			status: x.add,
		})
		if acc.worstStatus != x.expected {
			t.Errorf("expected %s, got %s", x.expected.Label, acc.worstStatus.Label)
		}
	}
}

func TestUpdateCheck(t *testing.T) {
	acc := newAccumulator()
	acc.add(result{
		status:  nagios.StatusCrit,
		message: "foo",
	})
	acc.add(result{
		status:  nagios.StatusWarn,
		message: "bar",
	})
	c := nagios.NewCheck()

	acc.updateCheck(c, func(ss []string) string {
		return strings.Join(ss, "|")
	})

	if c.Message != "foo|bar" {
		t.Error("unexpected message", c.Message)
	}

	if c.Status != nagios.StatusCrit {
		t.Error("unexpected status", c.Status)
	}
}
