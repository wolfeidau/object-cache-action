package main

import (
	"fmt"
	"log"

	actions "github.com/sethvargo/go-githubactions"
)

var (
	version = "dev"
)

func main() {

	p, err := getParams()
	if err != nil {
		log.Fatalf("failed to get params: %v", err)
	}

	fmt.Printf("::notice endpoint=%s\n", p.Endpoint)
	fmt.Printf("::notice key=%s\n", p.Key)
	fmt.Printf("::notice path=%s\n", p.Path)
	fmt.Printf("::notice restore-keys=%s\n", p.RestoreKeys)

	actions.SetOutput("cache-hit", "true")
}

type params struct {
	Key         string
	Path        string
	RestoreKeys string
	Endpoint    string
}

func getParams() (params, error) {
	return params{
		Key:         actions.GetInput("key"),
		Path:        actions.GetInput("path"),
		RestoreKeys: actions.GetInput("restore-keys"),
		Endpoint:    actions.GetInput("endpoint"),
	}, nil
}
