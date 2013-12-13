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
	engine.RegisterInputFactory(plugins.GetForwardInputFactory())
	inputFactory := engine.LookupInputFactory("forward")
	if inputFactory == nil {
		return
	}
	spawner := ik.NewSpawner()
	input, err := inputFactory.New(engine, map[string]string { "port": "24224" })
	if err != nil {
		println(err.Error())
		return
	}
	spawner.Spawn(input)
	spawner.PollMultiple([]ik.Spawnee { input })
}

// vim: sts=4 sw=4 ts=4 noet
