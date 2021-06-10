package main

import (
	"flag"
	"fmt"
	"github.com/luizfsnunes/assis/assis"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

func main() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveCmd.String("port", "8080", "Server port")
	pathServe := serveCmd.String("root", "", "HTML files")

	generate := flag.NewFlagSet("generate", flag.ExitOnError)
	generatePath := generate.String("path", "", "Project path")

	if len(os.Args) <= 1 {
		fmt.Println("no command supplied. expected: init, generate, serve")
		os.Exit(1)
	}

	start := time.Now()

	switch os.Args[1] {
	case "serve":
		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		server := assis.NewStaticServer(buildZap(), *port)
		if err := server.ListenAndServe(*pathServe); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
	case "generate":
		if err := generate.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		fmt.Println(generatePath)
		if err := generateSite(*generatePath); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Println("invalid command")
		os.Exit(1)
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %s to execute the command.", elapsed)

	os.Exit(0)
}

func buildZap() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	return logger
}

func generateSite(path string) error {
	logger := buildZap()

	config := assis.NewDefaultConfig(path)
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
