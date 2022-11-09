package main

import (
	"github.com/m10x/midi2key-ng/pkg/pkgGui"
)

const (
	VERSION_TOOL = "pre v1.0"
	VERSION_PREF = 1
)

func main() {
	pkgGui.Startup(VERSION_TOOL, VERSION_PREF)
}
