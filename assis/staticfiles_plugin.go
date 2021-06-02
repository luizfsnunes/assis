package assis

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
)

type StaticFilesPlugin struct {
	config     Config
	allowedExt []string
	logger     *zap.Logger
}

func NewStaticFilesPlugin(config Config, allowedExt []string, logger *zap.Logger) StaticFilesPlugin {
	return StaticFilesPlugin{
		config:     config,
		allowedExt: allowedExt,
		logger:     logger,
	}
}

func (s StaticFilesPlugin) AfterLoadFiles(files SiteFiles) error {
	s.logger.Info("Start static files copy")
	for _, container := range files {
		files := container.FilterExt(s.allowedExt)
		for _, file := range files {
			if err := GenerateDir(container.OutputFilename(file)); err != nil {
				return err
			}

			source, err := os.Open(container.FullFilename(file))
			if err != nil {
				return err
			}

			target, err := os.Create(container.OutputFilename(file))
			if err != nil {
				return err
			}

			_, err = io.Copy(target, source)
			if err != nil {
				return err
			}
			s.logger.Info(fmt.Sprintf("Source file: %s", container.FullFilename(file)))
			s.logger.Info(fmt.Sprintf("Target file: %s", container.OutputFilename(file)))
		}
	}
	s.logger.Info("Finished static files copy")
	return nil
}
