package assis

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type OS interface {
	ShouldGenerate(string) bool
}

type Windows struct {}

func (w Windows) ShouldGenerate(op string) bool  {
	return op == "REMOVE" || op == "CREATE"
}

type StaticServe struct {
	logger  *zap.Logger
	watcher *fsnotify.Watcher
	listen  func() error
	config  *Config
}

func NewStaticServer(config *Config, logger *zap.Logger, listen func() error) StaticServe {
	return StaticServe{
		logger: logger,
		config: config,
		listen: listen,
	}
}

func (s StaticServe) ListenAndServe() error {
	loggingHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.logger.Info(r.URL.Path)
			h.ServeHTTP(w, r)
		})
	}

	abs, _ := filepath.Abs(s.config.Output)

	http.Handle("/", loggingHandler(http.FileServer(http.Dir(abs))))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if s.listen != nil {
		go func() {
			if err := s.Watch(); err != nil {
				s.logger.Error(err.Error())
			}
		}()
	}

	go func() {
		if err := http.ListenAndServe(":"+s.config.Server.Port, nil); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("listen: %s\n", zap.Error(err))
		}
	}()

	s.logger.Info("Server Started")

	<-done
	s.logger.Info("Server Stopped")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	s.logger.Info("Server Exited Properly")

	return nil
}

func (s StaticServe) watchDir(path string, fi os.FileInfo, err error) error {
	if fi.Mode().IsDir() {
		return s.watcher.Add(path)
	}
	return nil
}

func (s StaticServe) Watch(os OS) error {
	s.watcher, _ = fsnotify.NewWatcher()
	defer s.watcher.Close()

	abs, err := filepath.Abs(s.config.Content)
	if err != nil {
		return err
	}

	if err := filepath.Walk(abs, s.watchDir); err != nil {
		return err
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			// watch for events
			case event := <-s.watcher.Events:
				if os.ShouldGenerate(event.Op.String()) {
					s.logger.Info("File update")
					if err := s.listen(); err != nil {
						s.logger.Info(err.Error())
					}
				}
			case err := <-s.watcher.Errors:
				s.logger.Error(err.Error())
			}
		}
	}()

	<-done
	return nil
}