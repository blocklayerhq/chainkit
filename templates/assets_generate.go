// +build ignore

package main

import (
	"log"

	"github.com/blocklayerhq/chainkit/templates"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(templates.Assets, vfsgen.Options{
		PackageName:  "templates",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
