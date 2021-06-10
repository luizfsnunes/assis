package assis

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestAssis_LoadFiles(t *testing.T) {
	cfg := NewDefaultConfig("./mock/_site")
	t.Run("with no plugins", func(t *testing.T) {
		assis := NewAssis(cfg, []interface{}{}, zaptest.NewLogger(t))
		err := assis.LoadFilesAsync()
		assert.NoError(t, err)
	})

	t.Run("with plugins", func(t *testing.T) {
		logger := zaptest.NewLogger(t)

		p := NewStaticFilesPlugin(cfg, []string{".js", ".png", ".jpg", ".jpeg", ".gif", ".css"}, logger)
		assis := NewAssis(NewDefaultConfig("./mock/_site"), []interface{}{p}, logger)
		err := assis.LoadFilesAsync()
		assert.NoError(t, err)
	})
}

func TestAssis_Generate(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := NewDefaultConfig("./mock/_site")
	p := NewStaticFilesPlugin(cfg, []string{".js", ".png", ".jpg", ".jpeg", ".gif", ".css"}, logger)

	assis := NewAssis(cfg, []interface{}{NewArticlePlugin(cfg, logger), NewHTMLPlugin(cfg, logger), p}, logger)
	assis.LoadFilesAsync()

	err := assis.Generate()
	assert.NoError(t, err)
}
