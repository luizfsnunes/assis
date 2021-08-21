package assis

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gammazero/workerpool"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"go.uber.org/zap"
)

type MinifyPlugin struct {
	config *Config
	minify     *minify.M
	logger     *zap.Logger
}

func NewMinifyPlugin(config *Config, logger *zap.Logger) *MinifyPlugin {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	return &MinifyPlugin{
		config: config,
		minify:     m,
		logger:     logger,
	}
}

func (m MinifyPlugin) allowedMediaType(filename string) string {
	return m.getPluginsConfig()[filepath.Ext(filename)]
}

func (m MinifyPlugin) AfterGeneratedFiles(files []string) error {
	m.logger.Info("Start minifying")
	wp := workerpool.New(2)
	maxJobs := 0
	for _, f := range files {
		f := f
		media := m.allowedMediaType(f)
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

func (m MinifyPlugin) getPluginsConfig() map[string]string {
	var plugincnfg = make(map[string]string)

	mediaTypes := fmt.Sprintf("%v", m.config.Plugins["minify_plugin"]["media_types"])
	cutMapping := strings.Replace(mediaTypes, "map[", "", -1)
	cutMapping = strings.Replace(cutMapping, "]", "", -1)


	for _, value := range strings.Split(cutMapping, " ") {
		mediaTypesSplit := strings.Split(value, ":")

		if len(mediaTypesSplit) > 1 {
			plugincnfg[mediaTypesSplit[0]] = mediaTypesSplit[1]
		}
	}
	return plugincnfg
}
