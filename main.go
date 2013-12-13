package main

import (
	"flag"
	"fmt"
	"github.com/moriyoshi/ik"
	"github.com/moriyoshi/ik/plugins"
	"log"
	"os"
	"path"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	logger := log.New(os.Stdout, "[ik] ", log.Lmicroseconds)

	var config_file string
	var help bool
	flag.StringVar(&config_file, "c", "/etc/fluent/fluent.conf", "config file path (default: /etc/fluent/fluent.conf)")
	flag.BoolVar(&help, "h", false, "show help")
	flag.Parse()

	if help || config_file == "" {
		usage()
	}

	scoreKeeper := ik.NewScoreKeeper()
	router := ik.NewFluentRouter()
	engine := ik.NewEngine(logger, scoreKeeper, router)
	engine.RegisterInputFactory(plugins.GetForwardInputFactory())
	engine.RegisterOutputFactory(plugins.GetStdoutOutputFactory())
	engine.RegisterOutputFactory(plugins.GetForwardOutputFactory())
	engine.RegisterOutputFactory(plugins.GetFileOutputFactory())

	engine.SetDefaultPort(router)

	spawner := ik.NewSpawner()

	var input ik.Input
	var err error

	dir, file := path.Split(config_file)
	opener := ik.DefaultOpener(dir)
	config, err := ik.ParseConfig(opener, file)
	if err != nil {
		println(err.Error())
		return
	}

	spawnees := make([]ik.Spawnee, 0)
	for _, v := range config.Root.Elems {
		switch v.Name {
		case "source":
			inputFactory := engine.LookupInputFactory(v.Attrs["type"])
			delete(v.Attrs, "type")
			if inputFactory == nil {
				return
			}
			input, err = inputFactory.New(engine, v.Attrs)
			if err != nil {
				logger.Fatal(err.Error())
				return
			}
			spawnees = append(spawnees, input)
			spawner.Spawn(input)
			logger.Printf("Input plugin loaded: ", inputFactory.Name())
		case "match":
			outputFactory := engine.LookupOutputFactory(v.Attrs["type"])
			output, err := outputFactory.New(engine, v.Attrs)
			router.AddRule(v.Args, output)
			if err != nil {
				logger.Fatal(err.Error())
				return
			}
			spawnees = append(spawnees, input)
			spawner.Spawn(output)
			logger.Printf("Output plugin loaded: %s, with Args '%s'", v.Name, v.Args)
        }
	}

	spawner.PollMultiple(spawnees)
}

// vim: sts=4 sw=4 ts=4 noet
