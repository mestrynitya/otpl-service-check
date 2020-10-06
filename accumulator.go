package main

import (
	"sync"

	"github.com/segfaultax/go-nagios"
)

type result struct {
	status  nagios.Status
	message string
	perf    []nagios.PerfData
}

type resultAccumulator struct {
	mu          sync.Mutex
	worstStatus nagios.Status
	results     []result
}

func newAccumulator() *resultAccumulator {
	return &resultAccumulator{
		worstStatus: nagios.StatusOK,
	}
}

func (a *resultAccumulator) add(r result) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.worstStatus = worseOf(a.worstStatus, r.status)
	a.results = append(a.results, r)
}

func (a *resultAccumulator) updateCheck(c *nagios.Check, formatter func([]string) string) {
	c.Status = a.worstStatus

	var msgs []string
	for _, r := range a.results {
		for _, p := range r.perf {
			c.AddPerfData(p)
		}
		msgs = append(msgs, r.message)
	}

	c.SetMessage(formatter(msgs))
}

func worseOf(a, b nagios.Status) nagios.Status {
	if statusToOrd(b) > statusToOrd(a) {
		return b
	}
	return a
}

func statusToOrd(s nagios.Status) int {
	// order: ok, unknown, warn, crit
	switch s {
	case nagios.StatusOK:
		return 0
	case nagios.StatusUnknown:
		return 1
	case nagios.StatusWarn:
		return 2
	case nagios.StatusCrit:
		return 3
	default:
		return -1
	}
}
