package main

import (
	"github.com/m10x/midi2key-ng/pkg/pkgGui"
)

const (
	VERSION_TOOL = "v1.1.1"
)

func main() {
	pkgGui.Startup(VERSION_TOOL)
}
