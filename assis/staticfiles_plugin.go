package assis

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
)

type StaticFilesPlugin struct {
	config     *Config
	logger     *zap.Logger
}

func NewStaticFilesPlugin(config *Config, logger *zap.Logger) *StaticFilesPlugin {
	return &StaticFilesPlugin{
		config:     config,
		logger:     logger,
	}
}

func (s StaticFilesPlugin) AfterLoadFiles(files SiteFiles) error {
	s.logger.Info("Start static files copy")

	s.logger.Info(fmt.Sprintf("Extensions enabled: %v", s.getPluginsConfig()))

	wp := workerpool.New(2)
	for _, container := range files {
		container := container
		wp.Submit(func() {
			if err := s.copyStaticFile(container); err != nil {
				s.logger.Error(err.Error())
			}
		})
	}
	wp.StopWait()
	s.logger.Info("Finished static files copy")
	return nil
}

func (s StaticFilesPlugin) copyStaticFile(container *FileContainer) error {
	files := container.FilterExt(s.getPluginsConfig())
	for _, file := range files {
		err := func() error {
			if err := GenerateDir(container.OutputFilename(file)); err != nil {
				return err
			}

			source, err := os.Open(container.FullFilename(file))
			defer func(source *os.File) {
				err := source.Close()
				if err != nil {

				}
			}(source)
			if err != nil {
				return err
			}

			target, err := os.Create(container.OutputFilename(file))
			defer func(target *os.File) {
				err := target.Close()
				if err != nil {

				}
			}(target)
			if err != nil {
				return err
			}

			if _, err = io.Copy(target, source); err != nil {
				return err
			}

			s.logger.Info(fmt.Sprintf("Source file: %s", container.FullFilename(file)))
			s.logger.Info(fmt.Sprintf("Target file: %s", container.OutputFilename(file)))
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s StaticFilesPlugin) getPluginsConfig() []string{
	var extensionList []string

	extensions := fmt.Sprintf("%v,", s.config.Plugins["static_files"]["extensions"])

	cutMapping := strings.Replace(extensions, "[",  "", -1)
	cutMapping = strings.Replace(cutMapping, "]", "", -1)
	cutMapping = strings.Replace(cutMapping, ",", "", -1)

	for _, value := range strings.Split(cutMapping, " ") {
		extensionList = append(extensionList, value)
	}

	return extensionList
}
