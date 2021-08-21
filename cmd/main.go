package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/luizfsnunes/assis/assis"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("no command supplied. expected: generate, serve")
		os.Exit(1)
	}

	serve := flag.NewFlagSet("serve", flag.ExitOnError)
	serveCfg := serve.String("config", "", "Config file")
	watch := serve.Bool("watch", false, "Watch files and hot-reload")

	generate := flag.NewFlagSet("generate", flag.ExitOnError)
	generateCfg := generate.String("config", "", "Config file")

	logger := buildZap()

	switch os.Args[1] {
	case "serve":
		if err := serve.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		config, err := assis.NewConfigFromFile(*serveCfg)
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

		config, err := assis.NewConfigFromFile(*generateCfg)
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

	plugins := assis.NewPluginRegistry(
		assis.NewArticlePlugin(config, logger),
		assis.NewHTMLPlugin(config, logger),
		assis.NewStaticFilesPlugin(config, logger),
		assis.NewMinifyPlugin(config, logger),
	)

	assisGenerator := assis.NewAssis(config, plugins, logger)
	if err := assisGenerator.LoadFilesAsync(); err != nil {
		return err
	}
	if err := assisGenerator.Generate(); err != nil {
		return err
	}
	return nil
}
