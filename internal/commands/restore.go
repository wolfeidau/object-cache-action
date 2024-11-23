package commands

import "context"

type RestoreCmd struct{}

func (cmd *RestoreCmd) Run(ctx context.Context, globals *Globals) error {
	return nil
}
