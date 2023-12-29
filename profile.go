package hokuchi

type Profile struct {
	ID         string            `json:"id"`
	Arch       string            `json:"arch"`
	Labels     map[string]string `json:"labels"`
	BootConfig BootConfig        `json:"boot_config"`
	Ignition   Ignition          `json:"ignition"`
}

type BootConfig struct {
	Kernel Kernel   `json:"kernel"`
	Images []Image  `json:"images"`
	Args   []string `json:"args"`
}

type Kernel struct {
	URI  string   `json:"uri"`
	Args []string `json:"args,omitempty"`
}

type Image struct {
	Name    string `json:"name,omitempty"`
	URI     string `json:"uri"`
}

type Ignition struct {
	Inline string `json:"inline,omitempty"`
	Source string `json:"source,omitempty"`
}
