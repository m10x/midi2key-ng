package main

import (
	"log"

	"github.com/m10x/midi2key-ng/pkg/pkgGui"
)

const (
	VERSION_TOOL = "v2.0.0"
)

func main() {
	log.Println("Hello world!")
	pkgGui.Startup(VERSION_TOOL)
}
