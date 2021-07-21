package assis

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

type PluginRegistry struct {
	render         []PluginRender
	generatedFiles []PluginGeneratedFiles
	loadFiles      []PluginLoadFiles
	customFunction []PluginCustomFunction
}

func NewPluginRegistry(plugins ...interface{}) PluginRegistry {
	r := PluginRegistry{
		render:         []PluginRender{},
		generatedFiles: []PluginGeneratedFiles{},
		loadFiles:      []PluginLoadFiles{},
		customFunction: []PluginCustomFunction{},
	}

	for _, plugin := range plugins {
		switch plugin := plugin.(type) {
		case PluginRender:
			r.render = append(r.render, plugin)
			break
		case PluginGeneratedFiles:
			r.generatedFiles = append(r.generatedFiles, plugin)
			break
		case PluginLoadFiles:
			r.loadFiles = append(r.loadFiles, plugin)
			break
		case PluginCustomFunction:
			r.customFunction = append(r.customFunction, plugin)
			break
		}
	}
	return r
}

type pluginDispatcher struct {
	registry PluginRegistry
}

func (r pluginDispatcher) DispatchPluginRender(assisTemplate AssisTemplate, siteFiles SiteFiles, templates Templates) error {
	for _, plugin := range r.registry.render {
		if err := plugin.OnRender(assisTemplate, siteFiles, templates); err != nil {
			return err
		}
	}
	return nil
}

func (r pluginDispatcher) DispatchPluginCustomFunction() map[string]interface{} {
	funcMap := make(map[string]interface{}, len(r.registry.customFunction))
	for _, plugin := range r.registry.customFunction {
		for name, fun := range plugin.OnRegisterCustomFunction() {
			funcMap[name] = fun
		}
	}
	return funcMap
}

func (r pluginDispatcher) DispatchPluginLoadFiles(files SiteFiles) error {
	for _, plugin := range r.registry.loadFiles {
		if err := plugin.AfterLoadFiles(files); err != nil {
			return err
		}
	}
	return nil
}

func (r pluginDispatcher) DispatchPluginGeneratedFiles(files []string) error {
	for _, plugin := range r.registry.generatedFiles {
		if err := plugin.AfterGeneratedFiles(files); err != nil {
			return err
		}
	}
	return nil
}
