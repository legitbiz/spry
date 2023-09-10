package cmds

import (
	"github.com/spf13/cobra"
)

func Init() cobra.Command {
	var rootCmd = cobra.Command{
		Use:   "spry",
		Short: "Spry schema generation and db management",
		Long:  "Commands to generate schema migrations and/or manage databases for spry",
	}
	var schemaCmd = GetActorSchema()
	rootCmd.AddCommand(&schemaCmd)
	return rootCmd
}
