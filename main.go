package main

import (
	"os"

	"github.com/akkuman/webkit-screenshot/wk"
)

func main() {
	loader := wk.NewLoader()

	url := os.Args[1]

	config := wk.NewScreenshotConfig(url)

	go loader.Screenshot(*config)

	loader.Exec()
}
