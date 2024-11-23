package commands

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	actions "github.com/sethvargo/go-githubactions"
)

type SaveCmd struct {
	Key         string `help:"Key for a cache entry." env:"INPUT_KEY"`
	Path        string `help:"Path list for a cache entry." env:"INPUT_PATH"`
	RestoreKeys string `help:"Restore keys list for a cache entry." env:"INPUT_RESTORE_KEYS"`
	Endpoint    string `help:"Endpoint for a cache entry." env:"INPUT_ENDPOINT"`
}

func (cmd *SaveCmd) Run(ctx context.Context, globals *Globals) error {

	fmt.Println("::notice saving cache")
	fmt.Printf("::notice endpoint=%s\n", cmd.Endpoint)
	fmt.Printf("::notice key=%s\n", cmd.Key)
	fmt.Printf("::notice path=%s\n", cmd.Path)
	fmt.Printf("::notice restore-keys=%s\n", cmd.RestoreKeys)

	if cmd.Endpoint == "http://localhost:8080" {
		return nil
	}

	token, err := actions.GetIDToken(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get id token: %w", err)
	}

	cacheURL, err := url.JoinPath(cmd.Endpoint, "cache", cmd.Key)
	if err != nil {
		return fmt.Errorf("failed to join path: %w", err)
	}

	// use the token to make a request to the cache service

	req, err := http.NewRequest("GET", cacheURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("cache-action/%s", globals.Version))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	actions.SetOutput("cache-hit", "true")

	return nil
}
