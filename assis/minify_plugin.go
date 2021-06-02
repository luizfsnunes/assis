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
	minify *minify.M
	logger *zap.Logger
}

func NewMinifyPlugin(logger *zap.Logger) MinifyPlugin {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	return MinifyPlugin{
		minify: m,
		logger: logger,
	}
}

func (m MinifyPlugin) AfterGeneratedFiles(files []string) error {
	m.logger.Info("Start minifying")
	mediaTypes := map[string]string{".html": "text/html", ".css": "text/css", ".js": "application/javascript"}
	for _, f := range files {
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		media := mediaTypes[filepath.Ext(f)]
		if media == "" {
			continue
		}

		b, err := m.minify.Bytes(media, content)
		if err != nil {
			return err
		}

		write, err := os.Create(f)
		if err != nil {
			return err
		}
		if _, err := write.WriteString(string(b)); err != nil {
			return err
		}
		write.Close()

		m.logger.Info(fmt.Sprintf("Minified: %s", f))
	}
	m.logger.Info("Finished minifying")
	return nil
}
