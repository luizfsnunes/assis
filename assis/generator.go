package assis

import (
	"html/template"

	"github.com/google/uuid"
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
	dispatcher    PluginDispatcher
	templates     Templates
	assisTemplate AssisTemplate
}

func NewGenerator(templates Templates, dispatcher PluginDispatcher) Generator {
	return SiteGenerator{
		dispatcher:    dispatcher,
		templates:     templates,
		assisTemplate: NewAssisTemplate(dispatcher.DispatchPluginCustomFunction()),
	}
}

func (h SiteGenerator) Render(siteFiles SiteFiles) error {
	return h.dispatcher.DispatchPluginRender(h.assisTemplate, siteFiles, h.templates)
}
