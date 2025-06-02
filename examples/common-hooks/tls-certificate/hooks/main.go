package main

import (
	"github.com/deckhouse/module-sdk/pkg/app"

	_ "tlscertificate/subfolder"
)

func main() {
	app.Run()
}
