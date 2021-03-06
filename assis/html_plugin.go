package assis

import (
	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
	"html/template"
	"path/filepath"
)

type HTMLPlugin struct {
	config *Config
	name   string
	logger *zap.Logger
}

func NewHTMLPlugin(config *Config, logger *zap.Logger) HTMLPlugin {
	return HTMLPlugin{config: config, name: "html", logger: logger}
}

func (h HTMLPlugin) OnRegisterCustomFunction() map[string]interface{} {
	return map[string]interface{}{
		"truncate": h.Truncate,
	}
}

func (h HTMLPlugin) Truncate(size int, str template.HTML) template.HTML {
	s := string(str)
	var numRunes = 0
	for index, _ := range s {
		numRunes++
		if numRunes > size {
			return template.HTML(s[:index] + "...")
		}
	}
	return str
}

func (h HTMLPlugin) OnRender(t AssisTemplate, siteFiles SiteFiles, templates Templates) error {
	h.logger.Info("Start HTML rendering")
	wp := workerpool.New(2)
	maxJobs := 0
	for _, container := range siteFiles {
		container := container
		wp.Submit(func() {
			if err := h.processContainer(container, t, templates); err != nil {
				h.logger.Error(err.Error())
			}
		})
		maxJobs += 1
		if maxJobs == 4 {
			maxJobs = 0
			wp.StopWait()
			wp = workerpool.New(2)
		}
	}
	wp.StopWait()
	h.logger.Info("Finished HTML rendering")
	return nil
}

func (h HTMLPlugin) processContainer(container *FileContainer, t AssisTemplate, templates Templates) error {
	files := container.FilterExt([]string{HTML})
	for _, file := range files {
		filename := filepath.ToSlash(container.FullFilename(file))

		allTemplates := append(templates.GetTemplatesByDir(filename), filename)
		targetTemplate, err := t.GetTemplate().ParseFiles(allTemplates...)

		err = func() error {
			target, err := CreateTargetFile(container.OutputFilename(file))
			defer target.Close()
			if err != nil {
				return err
			}

			if err = targetTemplate.ExecuteTemplate(target, "layout", nil); err != nil {
				return err
			}

			h.logger.Info("Rendered file to " + target.Name())
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
