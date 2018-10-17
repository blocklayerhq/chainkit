package templates

import (
	// Make vendoring happy.
	_ "github.com/shurcooL/vfsgen"
)

//go:generate go run -tags=dev assets_generate.go
