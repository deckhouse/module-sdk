package main

import (
	_ "basic-example-module/hooks"
	_ "basic-example-module/hooks/go-hooks"
	_ "basic-example-module/hooks/go-hooks/001-main-hook-in-subfolder"
	_ "basic-example-module/hooks/go-hooks/002_main_hook_with_bad_folder"
	_ "basic-example-module/hooks/main-hook-in-folder"
)
