package main

import (
	_ "basic-example-module/subfolder"
	_ "basic-example-module/subfolder/go-hooks"
	_ "basic-example-module/subfolder/go-hooks/001-main-hook-in-subfolder"
	_ "basic-example-module/subfolder/go-hooks/002_main_hook_with_bad_folder"
	_ "basic-example-module/subfolder/main-hook-in-folder"

	"github.com/deckhouse/module-sdk/pkg/app"
)

func main() {
	app.Run()
}
