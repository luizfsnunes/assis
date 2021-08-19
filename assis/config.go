package assis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type (
	Config struct {
		SiteRoot string                   `json:"site_root"`
		Plugins  map[string]PluginOptions `json:"plugins"`
		Output   string                   `json:"output"`
		Content  string                   `json:"content"`
		Template Template                 `json:"template"`
		Server   Server                   `json:"server"`
	}

	PluginOptions map[string]interface{}

	Template struct {
		Path     string `json:"path"`
		Partials string `json:"partials"`
		Layout   string `json:"layout"`
	}

	Server struct {
		Port string
	}
)

func NewConfig(configJson string) (Config, error) {
	var config Config
	if err := json.Unmarshal([]byte(configJson), &config); err != nil {
		return config, err
	}

	return config, nil
}

func NewConfigFromFile(configPath string) (*Config, error) {
	abs, err := filepath.Abs(configPath)
	pathFile := strings.Split(abs, "/")
	configFile := pathFile[len(pathFile)-1:]
	var mainPath string

	errPath := filepath.WalkDir(".", func(path string, d fs.DirEntry, e error) error {
		if !d.IsDir() && d.Name() == configFile[0] {
			mainPath = path
		}
		return nil
	})

	if errPath != nil {
		log.Fatal(errPath)
	}

	sitePath := strings.Split(mainPath, "/")

	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, err
	}

	var cfg *Config

	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	SetDefaultConfigs(cfg, sitePath[0])

	if err := cfg.validate(sitePath[0], configFile[0]); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c Config) validate(configFolder string, configFile string) error {
	if errFolder := checkSiteFolder(configFolder, c.SiteRoot); errFolder != nil {
		return errFolder
	}

	if errFile := checkConfigFile(configFolder, configFile); errFile != nil {
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

// TODO add these methods to Config struct
func checkSiteFolder(folder string, siteRoot string) error {

	if folder != siteRoot {
		return errors.New("you should define same of your folder that has in your site_root inside your config file")
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

	folder := strings.Split(content, "/")

	var emptyMessageError string

	messageDefaultVal := "\nyou must configure your config json like: \n" +
		"{\n  \"site_root\": \"your_site_root_folder\",\n  \"output\": \"name_your_output\",\n  " +
		"\"content\": \"content\",\n  " +
		"\"template\": {\n    \"path\": \"your_template_folder\",\n  " +
		"  \"partials\": \"your_template_folder/your_template_folder_partials\",\n  " +
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

	folder := strings.Split(template.Path, "/")

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

func checkConfigFile(folder string, cfgFile string) error {
	configFile, err := os.Stat(fmt.Sprintf("%s/%s", folder, cfgFile))

	if os.IsNotExist(err) {
		return errors.New(
			fmt.Sprintf("you must create a '%s' file inside your site folder to build project", cfgFile))
	}

	if configFile.IsDir() {
		return errors.New(fmt.Sprintf("%s must be a file", cfgFile))
	}

	return nil
}

func SetDefaultConfigs(config *Config, sitePath string) *Config {

	if len(config.SiteRoot) <= 0 {
		config.SiteRoot = sitePath
	}

	if len(config.Content) <= 0 {
		config.Content = fmt.Sprintf("%s/%s", sitePath, "content")
	} else {
		config.Content = fmt.Sprintf("%s/%s", sitePath, config.Content)
	}

	if len(config.Output) <= 0 {
		config.Output = fmt.Sprintf("%s/%s", sitePath, "output")
	} else {
		config.Output = fmt.Sprintf("%s/%s", sitePath, config.Output)
	}

	if len(config.Template.Path) <= 0 {
		config.Template.Path = fmt.Sprintf("%s/%s", sitePath, "template")
	} else {
		config.Template.Path = fmt.Sprintf("%s/%s", sitePath, config.Template.Path)
	}

	if len(config.Template.Partials) <= 0 {
		config.Template.Partials = fmt.Sprintf("%s/%s", sitePath, "partials")
	} else {
		config.Template.Partials = fmt.Sprintf("%s/%s", sitePath, config.Template.Partials)
	}

	if len(config.Template.Layout) <= 0 {
		config.Template.Layout = "index.html"
	}

	if len(config.Server.Port) <= 0 {
		config.Server.Port = "6780"
	}

	return config
}
