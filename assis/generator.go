package assis

import (
	"github.com/google/uuid"
	"html/template"
)

type AssisTemplate struct {
	funcMap template.FuncMap
}

func NewAssisTemplate(funcMap template.FuncMap) AssisTemplate {
	return AssisTemplate{funcMap: funcMap}
}

func (a AssisTemplate) GetTemplate() *template.Template {
	return template.New(uuid.New().String()).Funcs(a.funcMap)
}

type Generator interface {
	Render(files SiteFiles) error
}

type SiteGenerator struct {
	funcMap       template.FuncMap
	plugins       []interface{}
	templates     Templates
	assisTemplate AssisTemplate
}

func NewGenerator(templates Templates, plugins []interface{}) Generator {
	funcMap := make(map[string]interface{}, len(plugins))
	for _, plugin := range plugins {
		switch plugin := plugin.(type) {
		case PluginCustomFunction:
			for name, fun := range plugin.OnRegisterCustomFunction() {
				funcMap[name] = fun
			}
		}
	}

	return SiteGenerator{
		funcMap:       funcMap,
		plugins:       plugins,
		templates:     templates,
		assisTemplate: NewAssisTemplate(funcMap),
	}
}

func (h SiteGenerator) Render(siteFiles SiteFiles) error {
	for i := 0; i < len(h.plugins); i++ {
		switch plugin := h.plugins[i].(type) {
		case PluginRender:
			if err := plugin.OnRender(h.assisTemplate, siteFiles, h.templates); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
