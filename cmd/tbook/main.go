package main

import "github.com/bernardofernandezz/tui-ebook-reader/internal/cli"

// version é injetada no build: -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	cli.Execute(version)
}
