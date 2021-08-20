package assis_test

import (
	"github.com/luizfsnunes/assis/assis"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("run a formatted config with all options", func(t *testing.T) {
		config, err := assis.NewConfig(buildJSONMock())

		assert.NoError(t, err)
		assert.Equal(t, "out", config.Output)
		assert.Equal(t, "content", config.Content)

		expected := assis.PluginOptions{"extensions": ".svg,.js,.png,.jpg,.jpeg,.gif,.css"}
		assert.Equal(t, expected, config.Plugins["static_files"])

		expected = assis.PluginOptions{"media_types": "text/css,text/html,application/javascript"}
		assert.Equal(t, expected, config.Plugins["minify_plugin"])
	})
}

func buildJSONMock() string {
	return `{
  "output": "out",
  "content": "content",
  "template": {
    "path": "template",
    "base_template": "template/base.html",
    "partials": "template/partials"
  },
  "plugins": {
    "static_files": {
      "extensions": ".svg,.js,.png,.jpg,.jpeg,.gif,.css"
    },
    "minify_plugin": {
      "media_types": "text/css,text/html,application/javascript"
    }
  },
  "server": {
    "port": "8080"
  }
}`
}
