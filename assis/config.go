package assis

import (
	"path/filepath"
)

type Config struct {
	SiteRoot   string
	Content    string
	Output     string
	Template   string
	BaseLayout string
}

func NewDefaultConfig(siteRoot string) Config {
	siteRoot, _ = filepath.Abs(siteRoot)
	siteRoot = filepath.ToSlash(siteRoot)
	return Config{
		SiteRoot: siteRoot,
		Content:  siteRoot + "/content",
		Output:   siteRoot + "/out",
		Template: siteRoot + "/template",
	}
}
