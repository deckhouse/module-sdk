package main

import (
	"github.com/deckhouse/module-sdk/pkg/app"

	_ "dependency-example-module/subfolder"
)

func main() {
	app.Run()
}
