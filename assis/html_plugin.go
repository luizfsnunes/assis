package assis

import (
	"go.uber.org/zap"
	"html/template"
)

type HTMLPlugin struct {
	config Config
	name   string
	logger *zap.Logger
}

func NewHTMLPlugin(config Config, logger *zap.Logger) HTMLPlugin {
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
	allTemplates := templates
	for _, container := range siteFiles {
		files := container.FilterExt([]string{HTML})
		for _, file := range files {
			allTemplates = append(allTemplates, container.FullFilename(file))
			targetTemplate, err := t.GetTemplate().ParseFiles(allTemplates...)

			target, err := CreateTargetFile(container.OutputFilename(file))
			if err != nil {
				return err
			}

			if err = targetTemplate.ExecuteTemplate(target, "layout", nil); err != nil {
				return err
			}

			if err = target.Close(); err != nil {
				return err
			}

			h.logger.Info("Rendered file to " + target.Name())
		}
	}
	h.logger.Info("Finished HTML rendering")
	return nil
}