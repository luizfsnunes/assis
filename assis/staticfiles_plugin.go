package assis

import (
	"fmt"
	"io"
	"os"

	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
)

type StaticFilesPlugin struct {
	config     *Config
	allowedExt []string
	logger     *zap.Logger
}

func NewStaticFilesPlugin(config *Config, allowedExt []string, logger *zap.Logger) *StaticFilesPlugin {
	return &StaticFilesPlugin{
		config:     config,
		allowedExt: allowedExt,
		logger:     logger,
	}
}

func (s StaticFilesPlugin) AfterLoadFiles(files SiteFiles) error {
	s.logger.Info("Start static files copy")

	pluginList := s.getPluginsConfig()

	s.logger.Info(fmt.Sprintf("%v", pluginList))

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
	files := container.FilterExt(s.allowedExt)
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

func (s StaticFilesPlugin) getPluginsConfig() []PluginOptions {
	var pluginNameList []string
	var pluginList []PluginOptions

	for name, value := range s.config.Plugins {
		pluginNameList = append(pluginNameList, name)
		pluginList = append(pluginList, value)
	}

	s.logger.Info(fmt.Sprintf("Plugin List: %s", pluginNameList))
	return pluginList
}
