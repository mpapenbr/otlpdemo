package raw

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/raw/otlplog"
)

func NewRawOTLPCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "raw",
		Short: "collection of raw command to check otlp functionality",
		Long:  ``,
	}

	cmd.AddCommand(otlplog.NewLogEmitCommand())
	cmd.AddCommand(otlplog.NewOtelZapCommand())
	cmd.AddCommand(otlplog.NewZapContextCommand())
	cmd.AddCommand(otlplog.NewOwnLoggerCommand())
	return &cmd
}
