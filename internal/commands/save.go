package commands

import (
	"context"
	"fmt"

	actions "github.com/sethvargo/go-githubactions"
)

type SaveCmd struct {
	Key         string `help:"Key for a cache entry." env:"INPUT_KEY"`
	Path        string `help:"Path list for a cache entry." env:"INPUT_PATH"`
	RestoreKeys string `help:"Restore keys list for a cache entry." env:"INPUT_RESTORE_KEYS"`
	Endpoint    string `help:"Endpoint for a cache entry." env:"INPUT_ENDPOINT"`
}

func (cmd *SaveCmd) Run(ctx context.Context, globals *Globals) error {

	fmt.Printf("::notice endpoint=%s\n", cmd.Endpoint)
	fmt.Printf("::notice key=%s\n", cmd.Key)
	fmt.Printf("::notice path=%s\n", cmd.Path)
	fmt.Printf("::notice restore-keys=%s\n", cmd.RestoreKeys)

	actions.SetOutput("cache-hit", "true")

	return nil
}
