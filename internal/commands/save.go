package commands

import (
	"context"
	"fmt"
	"net/http"

	actions "github.com/sethvargo/go-githubactions"

	"github.com/wolfeidau/zipstash/pkg/archive"
	"github.com/wolfeidau/zipstash/pkg/client"
	"github.com/wolfeidau/zipstash/pkg/trace"
	"github.com/wolfeidau/zipstash/pkg/uploader"
)

type SaveCmd struct {
	Key      string `help:"Key for a cache entry." env:"INPUT_KEY"`
	Path     string `help:"Path list for a cache entry." env:"INPUT_PATH"`
	Endpoint string `help:"Endpoint for a cache entry." env:"INPUT_ENDPOINT"`
}

func (c *SaveCmd) Run(ctx context.Context, globals *Globals) error {
	ctx, span := trace.Start(ctx, "SaveCmd.Run")
	defer span.End()

	fmt.Println("::notice saving cache")
	fmt.Printf("::notice endpoint=%s\n", c.Endpoint)
	fmt.Printf("::notice key=%s\n", c.Key)
	fmt.Printf("::notice path=%s\n", c.Path)

	if c.Endpoint == "http://localhost:8080" {
		return nil
	}

	token, err := actions.GetIDToken(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get id token: %w", err)
	}

	err = c.save(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	actions.SetOutput("cache-hit", "true")

	return nil
}

func (c *SaveCmd) save(ctx context.Context, token string) error {
	ctx, span := trace.Start(ctx, "SaveCmd.save")
	defer span.End()

	paths, err := checkPath(c.Path)
	if err != nil {
		return fmt.Errorf("failed to check path: %w", err)
	}

	fileInfo, err := archive.BuildArchive(ctx, paths, c.Key)
	if err != nil {
		return fmt.Errorf("failed to build archive: %w", err)
	}

	cl, err := newClient(c.Endpoint, token)

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	createResp, err := cl.CreateCacheEntryWithResponse(ctx, "GitHubActions", client.CreateCacheEntryJSONRequestBody{
		CacheEntry: client.CacheEntry{
			Key:         c.Key,
			Compression: "zip",
			FileSize:    fileInfo.Size,
			Sha256sum:   fileInfo.Sha256sum,
			Paths:       paths,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create cache entry: %w", err)
	}

	if createResp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to create cache entry: %s", createResp.JSONDefault.Message)
	}

	upl := uploader.NewUploader(ctx, fileInfo.ArchivePath, createResp.JSON201.UploadInstructions, 20)

	etags, err := upl.Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	updateResp, err := cl.UpdateCacheEntryWithResponse(ctx, "GitHubActions", client.CacheEntryUpdateRequest{
		Id:             createResp.JSON201.Id,
		Key:            c.Key,
		MultipartEtags: etags,
	})
	if err != nil {
		return fmt.Errorf("failed to update cache entry: %w", err)
	}

	if updateResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to update cache entry: %s", updateResp.JSONDefault.Message)
	}

	return nil
}
