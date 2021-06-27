package main

import (
	"flag"
	"fmt"
	"github.com/luizfsnunes/assis/assis"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("no command supplied. expected: generate, serve")
		os.Exit(1)
	}

	serve := flag.NewFlagSet("serve", flag.ExitOnError)
	serveCfg := serve.String("config", "", "Config file")
	serveFolderCfg := serve.String("folder", "", "Site Folder")
	watch := serve.Bool("watch", false, "Watch files and hot-reload")

	generate := flag.NewFlagSet("generate", flag.ExitOnError)
	generateCfg := generate.String("config", "", "Config file")
	generateFolderCfg := generate.String("folder", "", "Site Folder")

	logger := buildZap()

	switch os.Args[1] {
	case "serve":
		if err := serve.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		config, err := assis.NewConfigFromFile(*serveFolderCfg, *serveCfg)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		var fn func() error
		if *watch == true {
			fn = func() error {
				return generateSite(config, logger)
			}
		}
		server := assis.NewStaticServer(config, logger, fn)
		if err = server.ListenAndServe(); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
	case "generate":
		if err := generate.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		config, err := assis.NewConfigFromFile(*generateFolderCfg, *generateCfg)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		if err = generateSite(config, logger); err != nil {
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

func generateSite(config *assis.Config, logger *zap.Logger) error {
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
