package assis

import (
	"fmt"
	"go.uber.org/zap"
	"io/fs"
	"path/filepath"
	"strings"
)

const HTML = ".html"
const MD = ".md"

type PluginRender interface {
	OnRender(AssisTemplate, SiteFiles, Templates) error
}

type PluginGeneratedFiles interface {
	AfterGeneratedFiles([]string) error
}

type PluginLoadFiles interface {
	AfterLoadFiles(SiteFiles) error
}

type PluginCustomFunction interface {
	OnRegisterCustomFunction() map[string]interface{}
}

type Templates []string

type File string

type FileContainer struct {
	entry       string
	files       []File
	outputPath  string
	contentPath string
}

func newFileContainer(outputPath, contentPath, entry string) *FileContainer {
	return &FileContainer{
		entry:       entry,
		files:       []File{},
		outputPath:  outputPath,
		contentPath: contentPath,
	}
}

func (f *FileContainer) AddOnEntry(file string) {
	f.files = append(f.files, File(file))
}

func (f *FileContainer) FullFilename(file File) string {
	return fmt.Sprintf("%s/%s", f.entry, string(file))
}

func (f *FileContainer) OutputFilename(file File) string {
	return strings.Replace(f.FullFilename(file), f.contentPath, f.outputPath, 1)
}

func (f *FileContainer) FilterExt(ext []string) []File {
	var filtered []File
	for _, file := range f.files {
		for _, e := range ext {
			if filepath.Ext(string(file)) == e {
				filtered = append(filtered, file)
			}
		}
	}
	return filtered
}

func (f *FileContainer) GetFile(filename string) string {
	for _, file := range f.files {
		if string(file) == filename {
			return f.FullFilename(file)
		}
	}
	return ""
}

type SiteFiles map[string]*FileContainer

func (s SiteFiles) Add(fc *FileContainer, entry string) {
	s[entry] = fc
}

func (s SiteFiles) Get(entry string) *FileContainer {
	return s[entry]
}

type Assis struct {
	config    Config
	templates Templates
	plugins   []interface{}
	container SiteFiles
	logger    *zap.Logger
}

func NewAssis(config Config, plugins []interface{}, logger *zap.Logger) Assis {
	logger.Info("Initializing generator")
	logger.Info(fmt.Sprintf("Content dir: %s", config.Content))
	logger.Info(fmt.Sprintf("Output dir: %s", config.Output))
	logger.Info(fmt.Sprintf("Template dir: %s", config.Template))

	return Assis{
		config:    config,
		plugins:   plugins,
		container: SiteFiles{},
		logger:    logger,
	}
}

func (a *Assis) LoadFiles() error {
	a.logger.Info("Run LoadFiles task")
	err := filepath.WalkDir(a.config.Template,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != HTML {
				return nil
			}

			a.templates = append(a.templates, filepath.ToSlash(path))
			a.logger.Info(fmt.Sprintf("Loaded template: %s", path))
			return nil
		})
	if err != nil {
		return err
	}

	err = filepath.WalkDir(a.config.Content,
		func(path string, d fs.DirEntry, err error) error {
			dirname := filepath.ToSlash(filepath.Dir(path))
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if a.container.Get(dirname) == nil {
				a.container.Add(newFileContainer(a.config.Output, a.config.Content, dirname), dirname)
			}
			a.container.Get(dirname).AddOnEntry(d.Name())
			return nil
		})
	if err != nil {
		return err
	}

	a.logger.Info("Run AfterLoadFiles")
	for _, plugin := range a.plugins {
		switch plugin := plugin.(type) {
		case PluginLoadFiles:
			if err := plugin.AfterLoadFiles(a.container); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (a *Assis) Generate() error {
	a.logger.Info("Run Generate task")
	generator := NewGenerator(a.templates, a.plugins)
	if err := generator.Render(a.container); err != nil {
		return err
	}

	var generated []string
	err := filepath.WalkDir(a.config.Output,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			generated = append(generated, filepath.ToSlash(path))
			return nil
		})
	if err != nil {
		return err
	}

	a.logger.Info("Run AfterGeneratedFiles")
	for _, plugin := range a.plugins {
		switch plugin := plugin.(type) {
		case PluginGeneratedFiles:
			if err := plugin.AfterGeneratedFiles(generated); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
