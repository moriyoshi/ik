package main

import (
	"github.com/moriyoshi/ik"
	"github.com/moriyoshi/ik/plugins"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[ik] ", log.Lmicroseconds)
	scoreKeeper := ik.NewScoreKeeper()
	engine := ik.NewEngine(logger, scoreKeeper)
	engine.RegisterInputFactory(plugins.GetStreamInputFactory())
	inputFactory := engine.LookupInputFactory("forward")
	if inputFactory == nil {
		return
	}
	input, err := inputFactory.New(engine, map[string]string { "port": "24224" })
	if err != nil {
		println(err.Error())
		return
	}
	input.Start()
	<-make(chan int)
}

// vim: sts=4 sw=4 ts=4 noet
