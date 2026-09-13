package main

import (
	"log"
	"os"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("uso: reader <arquivo.epub>")
	}

	book, err := epub.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if err := ui.Run(book); err != nil {
		log.Fatal(err)
	}
}
