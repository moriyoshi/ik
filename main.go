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
	engine := ik.NewEngine(logger, scoreKeeper)
	engine.RegisterInputFactory(plugins.GetForwardInputFactory())
	engine.RegisterOutputFactory(plugins.GetStdoutOutputFactory())

	var input ik.Input
	var err error

	dir, file := path.Split(config_file)
	opener := ik.DefaultOpener(dir)
	config, err := ik.ParseConfig(opener, file)
	if err != nil {
		println(err.Error())
		return
	}

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
			logger.Printf("Input plugin loaded: ", inputFactory.Name())
		case "match":
			outputFactory := engine.LookupOutputFactory(v.Attrs["type"])
			output, err := outputFactory.New(engine, v.Attrs)
			if err != nil {
				logger.Fatal(err.Error())
				return
			}
			engine.DefaultPort().RegisterOutput(output)
			logger.Printf("Onput plugin loaded: %s, with Args '%s'", v.Name, v.Args)
		}
	}

	input.Start()
	<-make(chan int)
}

// vim: sts=4 sw=4 ts=4 noet
