// Package config provides the read-only accessor that application code depends
// on to retrieve user/repository configuration.
//
// The interface is the stable contract; the backing implementation is expected
// to change. Today it is a hardcoded set of defaults ([Static]); later it will
// be a dynamic, file-backed store. Both satisfy [Config], so call sites never
// change when the backing store is swapped in.
//
// This package has no dependencies on internal/git, internal/model or
// internal/ui. It sits at the bottom of the import graph so every layer can
// depend on it without creating a cycle.
package config

// PullStrategy selects how `git pull` reconciles divergent upstream changes.
type PullStrategy string

const (
	PullMerge  PullStrategy = "--no-rebase"
	PullRebase PullStrategy = "--rebase"
	PullFFOnly PullStrategy = "--ff-only"
)

// String returns the git CLI flag for the strategy, satisfying fmt.Stringer.
func (s PullStrategy) String() string { return string(s) }

// Config is the accessor application code reads configuration through.
//
// Each setting is exposed as a typed method: callers get compile-time safety
// and values are validated/normalised at this boundary, so no downstream code
// has to interpret raw strings. Add a method here for each new setting.
type Config interface {
	// PullStrategy returns the configured strategy for `git pull`.
	PullStrategy() PullStrategy
}
