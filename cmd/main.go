package main

import (
	"encoding/json"
	"fmt"
	"github.com/luizfsnunes/assis/assis"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("no command supplied. expected: generate, serve")
		os.Exit(1)
	}

	var configFlag string
	if strings.HasPrefix(os.Args[2], "-config=") {
		configFlag = strings.Split(os.Args[2], "-config=")[1]
	}

	logger := buildZap()
	config, err := buildConfig(configFlag)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		fn := func() error {
			return generateSite(config, logger)
		}
		server := assis.NewStaticServer(config, logger, fn)
		if err := server.ListenAndServe(); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
	case "generate":
		if err := generateSite(config, logger); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Println("invalid command")
		os.Exit(1)
	}

	os.Exit(0)
}

func buildZap() *zap.Logger {
	logger, err := zap.NewDevelopment()
	defer logger.Sync()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return logger
}

func buildConfig(configFile string) (assis.Config, error) {
	abs, err := filepath.Abs(configFile)
	if err != nil {
		return assis.Config{}, err
	}

	b, err := ioutil.ReadFile(abs)
	if err != nil {
		return assis.Config{}, err
	}
	var cfg assis.Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return assis.Config{}, err
	}
	return cfg, nil
}

func generateSite(config assis.Config, logger *zap.Logger) error {
	plugins := []interface{}{
		assis.NewArticlePlugin(config, logger),
		assis.NewHTMLPlugin(config, logger),
		assis.NewStaticFilesPlugin(config, []string{".svg", ".js", ".png", ".jpg", ".jpeg", ".gif", ".css"}, logger),
		assis.NewMinifyPlugin(logger),
	}

	assisGenerator := assis.NewAssis(config, plugins, logger)
	if err := assisGenerator.LoadFilesAsync(); err != nil {
		return err
	}
	if err := assisGenerator.Generate(); err != nil {
		return err
	}
	return nil
}
