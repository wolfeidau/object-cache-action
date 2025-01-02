package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/zipstash/pkg/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/wolfeidau/object-cache-action/internal/commands"
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

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Logger()

	ctx := context.Background()

	tp, err := trace.NewProvider(ctx, "github.com/wolfeidau/object-cache-action", version)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create trace provider")
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
