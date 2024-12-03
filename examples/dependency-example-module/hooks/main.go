package main

import (
	_ "dependency-example-module/subfolder"

	"github.com/deckhouse/module-sdk/pkg/app"
)

func main() {
	app.Run()
}
