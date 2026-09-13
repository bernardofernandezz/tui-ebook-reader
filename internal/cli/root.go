package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "reader",
	Short: "Leitor de EPUB no terminal",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1) //cobra ja cuida da impressao do erro
	}
}
