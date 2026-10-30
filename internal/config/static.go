package config

import "time"

// Static is a hardcoded [Config] acts as temporary stand-in until a dynamic
// store is wired into git-manager
type Static struct {
	pullStrategy   PullStrategy
	autoFetchAfter time.Duration
}

// NewStatic returns a Static config populated with the built-in opinionated defaults.
func NewStatic() *Static {
	return &Static{
		pullStrategy:   PullMerge,
		autoFetchAfter: 12 * time.Hour,
	}
}

// PullStrategy implements [Config].
func (s *Static) PullStrategy() PullStrategy { return s.pullStrategy }

// AutoFetchAfter implements [Config].
func (s *Static) AutoFetchAfter() time.Duration { return s.autoFetchAfter }
