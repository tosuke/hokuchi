package flatcar

import (
	"fmt"
	"regexp"
)

type Key struct {
	channel string
	arch    string
	version string
}

func isChannelValid(channel string) bool {
	switch channel {
	case "stable", "beta", "alpha":
		return true
	}
	return false
}
func isArchValid(arch string) bool {
	switch arch {
	case "amd64", "arm64":
		return true
	}
	return false
}

var versionRegex = regexp.MustCompile("^\\d+\\.\\d+\\.\\d+$")

func isVersionValid(version string) bool {
	return versionRegex.MatchString(version)
}

func (k Key) String() string {
	return fmt.Sprintf("flatcar-%s-%s-%s", k.channel, k.arch, k.version)
}
func (k Key) baseURL() string {
	return fmt.Sprintf("https://%s.release.flatcar-linux.net/%s-usr/%s", k.channel, k.arch, k.version)
}
func (k Key) valid() bool {
	return isChannelValid(k.channel) && isArchValid(k.arch) && isVersionValid(k.version)
}