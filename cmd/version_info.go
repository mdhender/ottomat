package cmd

import "github.com/maloquacious/semver"

var (
	Version   = semver.Version{Major: 0, Minor: 1, Patch: 0, PreRelease: "alpha", Build: semver.Commit()}
	BuildDate string
	GitCommit string
)
