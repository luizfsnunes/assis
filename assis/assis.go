package assis

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const HTML = ".html"
const MD = ".md"

type Templates struct {
	cfg          *Config
	baseTemplate string
	partials     []string
	files        []string
	baseOrdered  []string
}

func NewTemplates(config *Config) Templates {
	return Templates{
		cfg:         config,
		partials:    []string{},
		files:       []string{},
		baseOrdered: []string{},
	}
}

func (t *Templates) orderBaseTemplate() {
	t.baseOrdered = []string{t.baseTemplate}
	t.baseOrdered = append(t.baseOrdered, t.partials...)
}

func (t *Templates) GetTemplatesByDir(fileToRender string) []string {
	fn := func(path string) int {
		return len(strings.Split(filepath.ToSlash(filepath.Dir(path)), "/"))
	}

	var out []string
	for _, tpl := range t.files {
		if fn(fileToRender) == fn(tpl) {
			out = append(out, tpl)
		}
	}
	return append(t.baseOrdered, out...)
}

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
	config     *Config
	templates  Templates
	dispatcher pluginDispatcher
	container  SiteFiles
	logger     *zap.Logger
}

func NewAssis(config *Config, registry PluginRegistry, logger *zap.Logger) Assis {
	logger.Info("Initializing generator")
	logger.Info(fmt.Sprintf("Content dir: %s", config.Content))
	logger.Info(fmt.Sprintf("Output dir: %s", config.Output))
	logger.Info(fmt.Sprintf("Template dir: %s", config.Template))

	return Assis{
		config:     config,
		dispatcher: pluginDispatcher{registry: registry},
		container:  SiteFiles{},
		logger:     logger,
		templates:  NewTemplates(config),
	}
}

func (a *Assis) LoadTemplates(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() || filepath.Ext(path) != HTML {
		return nil
	}

	if filepath.ToSlash(d.Name()) == a.config.Template.Layout {
		a.templates.baseTemplate = filepath.ToSlash(path)
		a.logger.Info(fmt.Sprintf("Loaded base template: %s", path))
		return nil
	}

	if strings.Contains(filepath.ToSlash(path), a.config.Template.Partials) {
		a.templates.partials = append(a.templates.partials, path)
		a.logger.Info(fmt.Sprintf("Loaded partial: %s", path))
		return nil
	}

	a.templates.files = append(a.templates.files, filepath.ToSlash(path))
	a.logger.Info(fmt.Sprintf("Loaded template: %s", path))
	return nil
}

func (a *Assis) LoadContent(path string, d fs.DirEntry, err error) error {
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
}

func (a *Assis) LoadFilesAsync() error {
	a.logger.Info("Run LoadFiles task")

	fatalErrors := make(chan error)
	wgDone := make(chan bool)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.WalkDir(a.config.Template.Path, a.LoadTemplates)
		if err != nil {
			fatalErrors <- err
		}
		a.templates.orderBaseTemplate()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.WalkDir(a.config.Content, a.LoadContent)
		if err != nil {
			fatalErrors <- err
		}
	}()

	go func() {
		wg.Wait()
		close(wgDone)
		close(fatalErrors)
	}()

	select {
	case <-wgDone:
		a.logger.Info("Run AfterLoadFiles")
		if err := a.dispatcher.DispatchPluginLoadFiles(a.container); err != nil {
			return err
		}
		break
	case err := <-fatalErrors:
		return err
	}

	return nil
}

func (a *Assis) Generate() error {
	a.logger.Info("Run Generate task")
	generator := NewGenerator(a.templates, a.dispatcher)
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
	if err := a.dispatcher.DispatchPluginGeneratedFiles(generated); err != nil {
		return err
	}
	return nil
}
