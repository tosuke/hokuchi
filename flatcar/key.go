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

func IsValidChannel(channel string) bool {
	switch channel {
	case "stable", "beta", "alpha":
		return true
	}
	return false
}
func IsValidArch(arch string) bool {
	switch arch {
	case "amd64", "arm64":
		return true
	}
	return false
}

var versionRegex = regexp.MustCompile("^\\d+\\.\\d+\\.\\d+$")

func IsValidVersion(version string) bool {
	if version == "current" {
		return true
	}
	return versionRegex.MatchString(version)
}

func (k Key) String() string {
	return fmt.Sprintf("flatcar-%s-%s-%s", k.channel, k.arch, k.version)
}
func (k Key) baseURL() string {
	return fmt.Sprintf("https://%s.release.flatcar-linux.net/%s-usr/%s", k.channel, k.arch, k.version)
}
func (k Key) valid() bool {
	return IsValidChannel(k.channel) && IsValidArch(k.arch) && IsValidVersion(k.version)
}

func (k Key) KernelKey() string {
	return k.String() + "-kernel"
}
func (k Key) InitrdKey() string {
	return k.String() + "-initrd"
}
