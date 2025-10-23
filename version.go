package ottomat

import (
	"github.com/maloquacious/semver"
)

var (
	version = semver.Version{Major: 0, Minor: 3, Patch: 4, PreRelease: "alpha", Build: semver.Commit()}
)

func Version() semver.Version {
	return version
}
