package assis

import (
	"log"
	"os"
	"path/filepath"
)

type CheckConfig interface {
	CheckConfigFolder()
	CheckConfigFile()
	//CheckBinFolder()
}

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

	checkPath struct {}
)

func NewCheckPath() *checkPath{
	return &checkPath{}
}

func (c *checkPath) CheckConfigFolder()  {
	configFolder, err := os.Stat("../_site")

	if os.IsNotExist(err) {
		log.Fatal("You must create a '_site' folder to build project.")
	}

	if !configFolder.IsDir() {
		log.Fatal("_site must be a folder.")
	}
}

func (c *checkPath) CheckConfigFile() {
	configFile, err := os.Stat("../_site/config.json")

	if os.IsNotExist(err){
		log.Fatal("You must create a 'config.json' file in '_site/' folder to build project.")
	}

	if configFile.IsDir() {
		log.Fatal("config.json must be a file.")
	}
}

/*func (c *checkPath) CheckBinFolder() {
	folder := "bin"

	_, err := os.Stat(folder)

	if os.IsNotExist(err){
		if err := os.Mkdir(folder, os.ModeDir); err != nil{
			log.Fatalf("Error to generate folder %v", err)
		}
	}
}*/

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
