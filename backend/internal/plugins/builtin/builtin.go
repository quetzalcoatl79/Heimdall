package builtin

// Import compiled-in plugins so their init() runs and they register themselves.

import (
	_ "github.com/nxo/engine/internal/plugins/builtin/captiveportal"
	_ "github.com/nxo/engine/internal/plugins/builtin/eviltwin"
	_ "github.com/nxo/engine/internal/plugins/builtin/healthcheck"
	_ "github.com/nxo/engine/internal/plugins/builtin/sample"
	_ "github.com/nxo/engine/internal/plugins/builtin/wifi"
)
