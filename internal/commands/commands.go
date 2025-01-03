package commands

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/wolfeidau/zipstash/pkg/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Globals struct {
	Version string
}

func newClient(endpoint, token string) (*client.ClientWithResponses, error) {

	httpClient := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	cl, err := client.NewClientWithResponses(endpoint, client.WithHTTPClient(httpClient), client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return cl, nil
}

func SplitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func checkPath(path string) ([]string, error) {
	paths := strings.Fields(path)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths provided")
	}

	return paths, nil
}
