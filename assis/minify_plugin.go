package assis

import (
	"fmt"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type MinifyPlugin struct {
	minify     *minify.M
	logger     *zap.Logger
	mediaTypes map[string]string
}

func NewMinifyPlugin(logger *zap.Logger) MinifyPlugin {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	return MinifyPlugin{
		minify:     m,
		logger:     logger,
		mediaTypes: map[string]string{".html": "text/html", ".css": "text/css", ".js": "application/javascript"},
	}
}

func (m MinifyPlugin) AllowedMediaType(filename string) string {
	return m.mediaTypes[filepath.Ext(filename)]
}

func (m MinifyPlugin) AfterGeneratedFiles(files []string) error {
	m.logger.Info("Start minifying")
	for _, f := range files {
		media := m.AllowedMediaType(f)
		if media == "" {
			continue
		}

		return func() error {
			content, err := ioutil.ReadFile(f)
			if err != nil {
				return err
			}

			b, err := m.minify.Bytes(media, content)
			if err != nil {
				return err
			}

			write, err := os.Create(f)
			defer write.Close()

			if err != nil {
				return err
			}
			if _, err := write.WriteString(string(b)); err != nil {
				return err
			}

			m.logger.Info(fmt.Sprintf("Minified: %s", f))
			return nil
		}()
	}
	m.logger.Info("Finished minifying")
	return nil
}
