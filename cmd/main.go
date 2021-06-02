package main

import (
	"flag"
	"fmt"
	assis2 "github.com/luizfsnunes/generator/assis"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveCmd.String("port", "8080", "Server port")
	pathServe := serveCmd.String("root", "", "HTML files")

	generate := flag.NewFlagSet("generate", flag.ExitOnError)
	generatePath := generate.String("path", "", "Project path")

	init := flag.NewFlagSet("init", flag.ExitOnError)
	initPath := init.String("path", "", "Project path")

	switch os.Args[1] {
	case "serve":
		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		if err := serveStatic(*pathServe, *port); err != nil {
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
	case "init":
		if err := init.Parse(os.Args[2:]); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		if err := initProject(*initPath); err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Println("invalid command")
		os.Exit(1)
	}
	os.Exit(0)
}

func generateSite(path string) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	config := assis2.NewDefaultConfig(path)
	plugins := []interface{}{
		assis2.NewArticlePlugin(config, logger),
		assis2.NewHTMLPlugin(config, logger),
		assis2.NewStaticFilesPlugin(config, []string{".svg", ".js", ".png", ".jpg", ".jpeg", ".gif", ".css"}, logger),
		assis2.NewMinifyPlugin(logger),
	}

	assis := assis2.NewAssis(config, plugins, logger)
	if err := assis.LoadFiles(); err != nil {
		return err
	}
	if err := assis.Generate(); err != nil {
		return err
	}
	return nil
}

func initProject(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	config := assis2.NewDefaultConfig(filepath.ToSlash(path))
	if err := os.Mkdir(config.Output, 0600); err != nil {
		return err
	}
	if err := os.Mkdir(config.Template, 0600); err != nil {
		return err
	}
	if err := os.Mkdir(config.Content, 0600); err != nil {
		return err
	}
	return nil
}

func serveStatic(path, port string) error {
	http.Handle("/", http.FileServer(http.Dir(path)))

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		return err
	}
	return nil
}
