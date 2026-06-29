package config

// Static is a hardcoded [Config] acts as temporary stand-in until a dynamic
// store is wired into git-manager
type Static struct {
	pullStrategy PullStrategy
}

// NewStatic returns a Static config populated with the built-in opinionated defaults.
func NewStatic() *Static {
	return &Static{
		pullStrategy: PullMerge,
	}
}

// PullStrategy implements [Config].
func (s *Static) PullStrategy() PullStrategy { return s.pullStrategy }
