package assis

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestHTMLGenerator_Render(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := NewDefaultConfig("./mock/_site")
	assis := NewAssis(config, nil, logger)
	assis.LoadFilesAsync()

	t.Run("generate HTML", func(t *testing.T) {
		gen := NewGenerator(assis.templates, []interface{}{NewHTMLPlugin(config, logger)})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})

	t.Run("generate Article", func(t *testing.T) {
		gen := NewGenerator(assis.templates, []interface{}{NewArticlePlugin(config, logger)})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})

	t.Run("test markdown custom function", func(t *testing.T) {
		gen := NewGenerator(assis.templates, []interface{}{NewArticlePlugin(config, logger), NewHTMLPlugin(config, logger)})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})

	t.Run("copy static files", func(t *testing.T) {
		p := NewStaticFilesPlugin(config, []string{".js", ".png", ".jpg", ".jpeg", ".gif", ".css"}, logger)
		gen := NewGenerator(assis.templates, []interface{}{p})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})
}
