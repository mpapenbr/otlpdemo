package raw

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/raw/config"
	"github.com/mpapenbr/otlpdemo/cmd/raw/otlplog"
)

func NewRawOTLPCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "raw",
		Short: "collection of raw command to check otlp functionality",
		Long:  ``,
	}
	cmd.PersistentFlags().StringVar(&config.OutputArg, "output", "stdout",
		"output destination (stdout, grpc)")
	cmd.AddCommand(otlplog.NewLogEmitCommand())
	cmd.AddCommand(otlplog.NewLogZapCommand())
	cmd.AddCommand(otlplog.NewLogOtelZapCommand())
	return &cmd
}
