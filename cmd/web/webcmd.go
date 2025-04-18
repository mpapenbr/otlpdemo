package web

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/web/httpclient"
)

func NewWebCommand() *cobra.Command {
	ret := cobra.Command{
		Use:   "web",
		Short: "collection of web commands",
		Long:  ``,
	}
	ret.AddCommand(httpclient.NewJSONPlaceholderCommand())
	return &ret
}
