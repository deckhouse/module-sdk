package main

import (
	_ "my-module/hooks"
	_ "my-module/hooks/go-hooks"
	_ "my-module/hooks/go-hooks/001-main-hook-in-subfolder"
	_ "my-module/hooks/go-hooks/002_main_hook_with_bad_folder"
	_ "my-module/hooks/main-hook-in-folder"
)
