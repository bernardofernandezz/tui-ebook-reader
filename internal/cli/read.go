package cli

import (
	"fmt"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/spf13/cobra"
)

var readChapter int

var readCmd = &cobra.Command{
	Use:   "read <arquivo.epub>",
	Short: "Imprime um capítulo no stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := epub.Open(args[0])
		if err != nil {
			return err
		}
		if readChapter < 1 || readChapter > len(book.Chapters) {
			return fmt.Errorf("capítulo %d não existe (o livro tem %d)", readChapter, len(book.Chapters))
		}
		fmt.Print(book.Text(book.Chapters[readChapter-1]))
		return nil
	},
}

func init() {
	readCmd.Flags().IntVarP(&readChapter, "cap", "c", 1, "número do capítulo")
	rootCmd.AddCommand(readCmd)
}
