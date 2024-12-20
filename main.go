package main

import (
	"context"
	"log"

	"github.com/alecthomas/kong"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/wolfeidau/object-cache-action/internal/commands"
	"github.com/wolfeidau/object-cache-action/internal/trace"
)

var (
	version = "dev"
	cli     struct {
		Version kong.VersionFlag
		Save    commands.SaveCmd    `cmd:"" help:"save files."`
		Restore commands.RestoreCmd `cmd:"" help:"restore files."`
	}
)

func main() {

	ctx := context.Background()

	tp, err := trace.NewProvider(ctx, "github.com/wolfeidau/object-cache-service", version)
	if err != nil {
		log.Fatalf("failed to create trace provider: %v", err)
	}
	defer func() {
		_ = tp.Shutdown(ctx)
	}()

	var span oteltrace.Span
	ctx, span = trace.Start(ctx, "object-cache-service")
	defer span.End()

	cmd := kong.Parse(&cli,
		kong.Vars{
			"version": version,
		},
		kong.BindTo(ctx, (*context.Context)(nil)))
	err = cmd.Run(&commands.Globals{Version: version})
	span.RecordError(err)
	cmd.FatalIfErrorf(err)
}
