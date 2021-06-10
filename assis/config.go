package assis

import (
	"path/filepath"
)

type (
	Config struct {
		SiteRoot string `json:"site_root"`
		Output   string `json:"output"`
		Content  string `json:"content"`
		Template Template
	}

	Template struct {
		Path     string `json:"path"`
		Partials string `json:"partials"`
		Layout   string `json:"layout"`
	}
)

func NewDefaultConfig(siteRoot string) Config {
	siteRoot, _ = filepath.Abs(siteRoot)
	siteRoot = filepath.ToSlash(siteRoot)
	return Config{
		SiteRoot: siteRoot,
		Content:  siteRoot + "/content",
		Output:   siteRoot + "/out",
		Template: Template{
			Path:     siteRoot + "/template",
			Partials: siteRoot + "/template/partials",
			Layout:   "layout.html",
		},
	}
}
