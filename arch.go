package hokuchi

func NormalizeArch(arch string) string {
	switch arch {
	case "arm64", "aarch64":
		return "arm64"
	case "amd64", "x86_64", "intel64":
		return "amd64"
	default:
		return arch
	}
}
