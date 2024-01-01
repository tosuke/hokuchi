package profile

type Profile struct {
	ID       string            `json:"id"`
	Arch     string            `json:"arch"`
	Labels   map[string]string `json:"labels"`
	Boot     Boot              `json:"boot"`
	Ignition Ignition          `json:"ignition"`
}

type Boot struct {
	Flatcar *Flatcar `json:"flatcar"`
}

type Flatcar struct {
	Channel string   `json:"channel"`
	Version string   `json:"version"`
	Args    []string `json:"args"`
}

type Ignition struct {
	Inline string `json:"inline,omitempty"`
	Source string `json:"source,omitempty"`
}

func (p Profile) ResourceSpecs() []ResourceSpec {
	var rs []ResourceSpec
	if fc := p.Boot.Flatcar; fc != nil {
		rs = append(rs, ResourceSpec{Flatcar: &FlatcarResourceSpec{
			Arch:    p.Arch,
			Channel: fc.Channel,
			Version: fc.Version,
		}})
	}

	return rs
}
