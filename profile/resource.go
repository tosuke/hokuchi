package profile

import "github.com/tosuke/hokuchi/flatcar"

type ResourceSpec struct {
	Flatcar *FlatcarResourceSpec `json:"flatcar,omitempty"`
	HTTP    *HTTPResourceSpec    `json:"http,omitempty"`
}

type FlatcarResourceSpec struct {
	Arch    string `json:"arch"`
	Channel string `json:"channel"`
	Version string `json:"version"`
}

type HTTPResourceSpec struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256,omitempty"`
}

func (rs ResourceSpec) Valid() bool {
	has := false

	if rs.Flatcar != nil {
		if !flatcar.IsValidArch(rs.Flatcar.Arch) || !flatcar.IsValidChannel(rs.Flatcar.Channel) || !flatcar.IsValidVersion(rs.Flatcar.Version) {
			return false
		}
		has = true
	}

	if rs.HTTP != nil {
		if has {
			return false
		}
		has = true
	}

	return has
}
