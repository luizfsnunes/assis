package assis

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func (c Config) validate(configFolder string) error {

	if errFolder := checkSiteFolder(configFolder, c.SiteRoot); errFolder != nil {
		return errFolder
	}

	if errFile := checkConfigFile(configFolder); errFile != nil {
		return errFile
	}

	if errContent := checkConfigFolders(c.Content, "content"); errContent != nil {
		return errContent

	}

	if errOutput := checkConfigFolders(c.Output, "output"); errOutput != nil {
		return errOutput
	}

	if errTemplates := checkConfigTemplate(c.Template); errTemplates != nil {
		return errTemplates
	}

	if errServer := checkConfigServer(c.Server); errServer != nil {
		return errServer
	}

	return nil
}

func checkSiteFolder(folder string, siteRoot string) error {

	if len(siteRoot) == 0 {
		return errors.New("you must define your site_root in your config.json")
	}

	if folder != siteRoot {
		return errors.New("you should define same of your folder that has in your site_root(config.json)")
	}

	configFolder, err := os.Stat(folder)

	if os.IsNotExist(err) {
		return errors.New("you must create a folder with the name of your site config to build project")
	}

	if !configFolder.IsDir() {
		return errors.New("try use other name to you site, or make it a folder")
	}

	return nil
}

func checkConfigFolders(content string, configType string) error {

	folder := strings.Split(content,"/")

	var emptyMessageError string

	messageDefaultVal := "\nyou must configure your config json like: \n" +
		"{\n  \"site_root\": \"example\",\n  \"output\": \"out\",\n  \"content\": \"content\",\n  " +
		"\"template\": {\n    \"path\": \"template\",\n    \"partials\": \"template/partials\",\n  " +
		"  \"layout\": \"base.html\"\n  },\n  \"server\": {\n    \"port\": \"8080\"\n  }\n}"

	switch configType {
	case "output":
		emptyMessageError = "you must define your output in your config.json"
	case "content":
		emptyMessageError = "you must define your content in your config.json"
	default:
		return errors.New(messageDefaultVal)
	}

	if folder[1] == "" && len(folder[1]) == 0 {
		return errors.New(emptyMessageError)
	}

	contentFolder, err := os.Stat(content)

	if os.IsNotExist(err) {
		return errors.New(
			fmt.Sprintf("you must create a folder %s inside your folder %s to build project", folder[1], folder[0]))
	}

	if !contentFolder.IsDir() {
		return errors.New(fmt.Sprintf("try use name '%s' or make it a folder", folder[1]))
	}

	return nil
}

func checkConfigTemplate(template Template) error {

	folder := strings.Split(template.Path,"/")

	if folder[1] == "" && len(folder[1]) == 0 {
		return errors.New("you must define your template path in your config.json")
	}

	if len(template.Partials) == 0 {
		return errors.New("you must define your template partials path in your config.json")
	}

	if len(template.Layout) == 0 {
		return errors.New("you must define your template layout in your config.json")
	}

	contentFolder, err := os.Stat(template.Path)

	if os.IsNotExist(err) {
		return errors.New(
			fmt.Sprintf("you must create a folder '%s' inside your folder %s to build project", folder[1], folder[0]))
	}

	if !contentFolder.IsDir() {
		return errors.New(fmt.Sprintf("try use name '%s' or make it a folder", folder[1]))
	}

	return nil
}

func checkConfigServer(server Server) error {

	if len(server.Port) > 4 {
		return errors.New("you must define a port of your server with more 4 characters in your config.json")
	}

	return nil
}


func checkConfigFile(folder string) error {
	configFile, err := os.Stat(fmt.Sprintf("%s/config.json", folder))

	if os.IsNotExist(err) {
		return errors.New("you must create a 'config.json' file inside your site folder to build project")
	}

	if configFile.IsDir() {
		return errors.New("config.json must be a file")
	}

	return nil
}

func NewConfigFromFile(cfgFolder string, configPath string) (*Config, error) {
	abs, err := filepath.Abs(fmt.Sprintf("%s/%s", cfgFolder, configPath))
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	var cfg *Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	cfg.Content = fmt.Sprintf("%s/%s", cfgFolder, cfg.Content)
	cfg.Output = fmt.Sprintf("%s/%s", cfgFolder, cfg.Output)
	cfg.Template.Path = fmt.Sprintf("%s/%s", cfgFolder, cfg.Template.Path)
	cfg.Template.Partials = fmt.Sprintf("%s/%s", cfgFolder, cfg.Template.Partials)

	if err := cfg.validate(cfgFolder); err != nil {
		return nil, err
	}

	return cfg, nil
}


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
