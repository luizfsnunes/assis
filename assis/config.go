package assis

import (
	"path/filepath"
)

type (
	Config struct {
		SiteRoot string   `json:"site_root"`
		Output   string   `json:"output"`
		Content  string   `json:"content"`
		Template Template `json:"template"`
		Server   Server   `json:"server"`
	}

	Template struct {
		Path     string `json:"path"`
		Partials string `json:"partials"`
		Layout   string `json:"layout"`
	}

	Server struct {
		Port string
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
		Server: Server{
			Port: "8080",
		},
	}
}
