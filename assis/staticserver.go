package assis

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type StaticServe struct {
	logger *zap.Logger
	port   string
}

func NewStaticServer(logger *zap.Logger, port string) StaticServe {
	return StaticServe{
		logger: logger,
		port:   ":" + port,
	}
}

func (s StaticServe) ListenAndServe(path string) error {
	loggingHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.logger.Info(r.URL.Path)
			h.ServeHTTP(w, r)
		})
	}

	http.Handle("/", loggingHandler(http.FileServer(http.Dir(path))))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(s.port, nil); err != nil && err != http.ErrServerClosed {
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
