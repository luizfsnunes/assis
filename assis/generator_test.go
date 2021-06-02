package assis

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHTMLGenerator_Render(t *testing.T) {
	config := NewDefaultConfig("./mock/_site")
	assis := NewAssis(config, nil)
	assis.LoadFiles()

	t.Run("generate HTML", func(t *testing.T) {
		gen := NewGenerator(assis.templates, []interface{}{NewHTMLPlugin(config)})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})

	t.Run("generate Article", func(t *testing.T) {
		gen := NewGenerator(assis.templates, []interface{}{NewArticlePlugin(config)})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})

	t.Run("test markdown custom function", func(t *testing.T) {
		gen := NewGenerator(assis.templates, []interface{}{NewArticlePlugin(config), NewHTMLPlugin(config)})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})

	t.Run("copy static files", func(t *testing.T) {
		p := NewStaticFilesPlugin(config, []string{".js", ".png", ".jpg", ".jpeg", ".gif", ".css"})
		gen := NewGenerator(assis.templates, []interface{}{p})
		err := gen.Render(assis.container)
		assert.NoError(t, err)
	})
}
