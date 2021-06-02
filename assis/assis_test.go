package assis

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssis_LoadFiles(t *testing.T) {
	cfg := NewDefaultConfig("./mock/_site")
	t.Run("with no plugins", func(t *testing.T) {
		assis := NewAssis(cfg, nil)
		err := assis.LoadFiles()
		assert.NoError(t, err)
	})

	t.Run("with plugins", func(t *testing.T) {
		p := NewStaticFilesPlugin(cfg, []string{".js", ".png", ".jpg", ".jpeg", ".gif", ".css"})
		assis := NewAssis(NewDefaultConfig("./mock/_site"), []interface{}{p})
		err := assis.LoadFiles()
		assert.NoError(t, err)
	})
}

func TestAssis_Generate(t *testing.T) {
	cfg := NewDefaultConfig("./mock/_site2")
	p := NewStaticFilesPlugin(cfg, []string{".js", ".png", ".jpg", ".jpeg", ".gif", ".css"})

	assis := NewAssis(cfg, []interface{}{NewArticlePlugin(cfg), NewHTMLPlugin(cfg), p})
	assis.LoadFiles()

	err := assis.Generate()
	assert.NoError(t, err)
}
