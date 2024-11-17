package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
)

var (
	version = "dev"

	cli struct {
		Debug        bool `help:"Enable debug mode."`
		Version      kong.VersionFlag
		Key          string `help:"Key for a cache entry." env:"INPUT_KEY"`
		Path         string `help:"Path list for a cache entry." env:"INPUT_PATH"`
		RestoreKeys  string `help:"Restore keys list for a cache entry." env:"INPUT_RESTORE_KEYS"`
		GitHubOutput string `help:"Path list for a cache entry." env:"GITHUB_OUTPUT" type:"path"`
	}
)

func main() {
	kong.Parse(&cli,
		kong.Vars{
			"version": version,
		})

	fmt.Printf("::notice %s\n", cli.Key)

	f, err := os.OpenFile(cli.GitHubOutput, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("failed to open output file: %v", err)
	}

	defer f.Close()

	_, err = f.WriteString("true")
	if err != nil {
		log.Fatalf("failed to write to output file: %v", err)
	}

}
