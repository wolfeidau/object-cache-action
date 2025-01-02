package commands

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/klauspost/compress/zip"
	"github.com/rs/zerolog/log"
	actions "github.com/sethvargo/go-githubactions"
	"github.com/wolfeidau/quickzip"
	"github.com/wolfeidau/zipstash/pkg/archive"
	"github.com/wolfeidau/zipstash/pkg/downloader"
	"github.com/wolfeidau/zipstash/pkg/trace"
	"go.opentelemetry.io/otel/attribute"
)

type RestoreCmd struct {
	Key      string `help:"Key for a cache entry." env:"INPUT_KEY"`
	Path     string `help:"Path list for a cache entry." env:"INPUT_PATH"`
	Endpoint string `help:"Endpoint for a cache entry." env:"INPUT_ENDPOINT"`
}

func (c *RestoreCmd) Run(ctx context.Context, globals *Globals) error {
	ctx, span := trace.Start(ctx, "RestoreCmd.Run")
	defer span.End()

	fmt.Println("::notice restore cache")
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

	return c.restore(ctx, token)
}

func (c *RestoreCmd) restore(ctx context.Context, token string) error {
	ctx, span := trace.Start(ctx, "RestoreCmd.restore")
	defer span.End()

	cl, err := newClient(c.Endpoint, token)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	getEntryResp, err := cl.GetCacheEntryByKeyWithResponse(ctx, "GitHubActions", c.Key)
	if err != nil {
		return fmt.Errorf("failed to get cache entry: %w", err)
	}

	// TODO: handle alternate restore keys
	if getEntryResp.StatusCode() == http.StatusNotFound {
		log.Warn().Msg("cache entry not found")
		actions.SetOutput("cache-hit", "")
		return nil
	}

	if getEntryResp.JSON200 == nil {
		return fmt.Errorf("failed to get cache entry: %s", getEntryResp.Status())
	}

	log.Info().Any("cache entry", getEntryResp.JSON200).Msg("cache entry")

	downloads, err := downloader.NewDownloader(getEntryResp.JSON200.DownloadInstructions, 20).Download(ctx)
	if err != nil {
		return fmt.Errorf("failed to download cache entry: %w", err)
	}

	slices.SortFunc(downloads, func(a, b downloader.DownloadedFile) int {
		return cmp.Compare(a.Part, b.Part)
	})

	for _, d := range downloads {
		log.Info().Any("download", d).Msg("download")
	}

	zipFile, zipFileLen, err := combineParts(ctx, downloads)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer zipFile.Close()

	log.Info().Int64("zipFileLen", zipFileLen).Str("name", zipFile.Name()).Msg("zip file len")

	paths, err := checkPath(c.Path)
	if err != nil {
		return fmt.Errorf("failed to check path: %w", err)
	}

	err = restoreFiles(ctx, zipFile, zipFileLen, paths)
	if err != nil {
		return fmt.Errorf("failed to restore files: %w", err)
	}

	// cleanup zip file
	defer os.Remove(zipFile.Name())

	return nil
}

func restoreFiles(ctx context.Context, zipFile *os.File, zipFileLen int64, paths []string) error {
	_, span := trace.Start(ctx, "restoreFiles")
	defer span.End()
	extract, err := quickzip.NewExtractorFromReader(zipFile, zipFileLen)
	if err != nil {
		return fmt.Errorf("failed to create extractor: %w", err)
	}

	mappings, err := archive.PathsToMappings(paths)
	if err != nil {
		return fmt.Errorf("failed to create mappings: %w", err)
	}

	err = extract.ExtractWithPathMapper(ctx, func(file *zip.File) (string, error) {
		for _, mapping := range mappings {
			if strings.HasPrefix(file.Name, mapping.RelativePath) {
				return filepath.Join(mapping.Chroot, file.Name), nil
			}
		}

		return "", fmt.Errorf("failed to find path mapping for: %s", file.Name)
	})
	if err != nil {
		return fmt.Errorf("failed to extract zip file: %w", err)
	}
	return nil
}

// pass in a list of paths and turn them into a zip file stream to enable extraction
func combineParts(ctx context.Context, downloads []downloader.DownloadedFile) (*os.File, int64, error) {
	_, span := trace.Start(ctx, "combineParts")
	defer span.End()

	zipFile, err := os.CreateTemp("", "zipstash-download-*.zip")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create temp file: %w", err)
	}

	zipFileLen := int64(0)

	for _, d := range downloads {
		n, err := appendToFile(zipFile, d.FilePath)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to write file: %w", err)
		}
		zipFileLen += n
	}

	span.SetAttributes(attribute.Int64("zipFileLen", zipFileLen))

	return zipFile, zipFileLen, nil
}

func appendToFile(f *os.File, path string) (int64, error) {
	pf, err := os.Open(filepath.Clean(path))
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer pf.Close()

	n, err := io.Copy(f, pf)
	if err != nil {
		return 0, fmt.Errorf("failed to copy file: %w", err)
	}

	defer os.Remove(path)

	return n, nil
}
