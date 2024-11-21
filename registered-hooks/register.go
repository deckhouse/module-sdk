// nolint: revive
package registeredhooks

import (
	_ "github.com/deckhouse/module-sdk/somemodule/hooks"
	_ "github.com/deckhouse/module-sdk/somemodule/hooks/001_ensure_crd"
	_ "github.com/deckhouse/module-sdk/somemodule/hooks/go-hooks/001_ensure_crd"
	_ "github.com/deckhouse/module-sdk/somemodule/hooks/go-hooks/002_ensure_crd_2"
)
