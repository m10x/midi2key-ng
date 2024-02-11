package main

import (
	"github.com/m10x/midi2key-ng/pkg/pkgGui"
	"log"
)

const (
	VERSION_TOOL = "v1.1.1"
)

func main() {
	log.Println("Hello world!")
	pkgGui.Startup(VERSION_TOOL)
}
