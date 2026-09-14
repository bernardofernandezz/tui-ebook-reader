package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <termo>",
	Short: "Busca livros no Project Gutenberg",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		books, err := searchGutenberg(query)
		if err != nil {
			return err
		}
		if len(books) == 0 {
			fmt.Printf("nenhum resultado para %q\n", query)
			return nil
		}

		for _, b := range books {
			fmt.Printf("%6d  %s — %s\n", b.ID, b.Title, b.author())
		}
		fmt.Printf("\n%d resultados · use: tbook download <id>\n", len(books))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
