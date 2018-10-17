// +build dev

package templates

import "net/http"

// Assets contains the project's assets.
var Assets http.FileSystem = http.Dir("src")
