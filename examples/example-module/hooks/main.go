package main

import (
	"github.com/deckhouse/module-sdk/pkg/app"

	_ "example-module/subfolder"
)

func main() {
	app.Run()
}
