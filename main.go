package main

import (
	"github.com/moriyoshi/ik"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "lk", log.Lmicroseconds)
	ik.Server(logger, "")
}

// vim: sts=4 sw=4 ts=4 noet
