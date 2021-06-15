package assis

import (
	"fmt"
	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
	"io"
	"os"
)

type StaticFilesPlugin struct {
	config     *Config
	allowedExt []string
	logger     *zap.Logger
}

func NewStaticFilesPlugin(config *Config, allowedExt []string, logger *zap.Logger) StaticFilesPlugin {
	return StaticFilesPlugin{
		config:     config,
		allowedExt: allowedExt,
		logger:     logger,
	}
}

func (s StaticFilesPlugin) AfterLoadFiles(files SiteFiles) error {
	s.logger.Info("Start static files copy")

	wp := workerpool.New(2)
	maxJobs := 0
	for _, container := range files {
		container := container
		wp.Submit(func() {
			if err := s.copyStaticFile(container); err != nil {
				s.logger.Error(err.Error())
			}
		})
		maxJobs += 1
		if maxJobs == 4 {
			maxJobs = 0
			wp.StopWait()
			wp = workerpool.New(2)
		}
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
			defer source.Close()
			if err != nil {
				return err
			}

			target, err := os.Create(container.OutputFilename(file))
			defer target.Close()
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
