package main

import (
	_ "example-module/hooks"

	"github.com/deckhouse/module-sdk/pkg/app"
)

func main() {
	app.Run()
}
