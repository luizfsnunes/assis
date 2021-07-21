package assis

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gammazero/workerpool"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"go.uber.org/zap"
)

type MinifyPlugin struct {
	minify     *minify.M
	logger     *zap.Logger
	mediaTypes map[string]string
}

func NewMinifyPlugin(logger *zap.Logger) *MinifyPlugin {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	return &MinifyPlugin{
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
	wp := workerpool.New(2)
	maxJobs := 0
	for _, f := range files {
		f := f
		media := m.AllowedMediaType(f)
		if media == "" {
			continue
		}
		wp.Submit(func() {
			if err := m.minifyFiles(f, media); err != nil {
				m.logger.Error(err.Error())

				maxJobs += 1
				if maxJobs == 8 {
					maxJobs = 0
					wp.StopWait()
					wp = workerpool.New(2)
				}
			}
		})
	}
	wp.StopWait()
	m.logger.Info("Finished minifying")
	return nil
}

func (m MinifyPlugin) minifyFiles(f, media string) error {
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
}
