package main

import (
	"flag"
	"fmt"
	"github.com/moriyoshi/ik"
	"github.com/moriyoshi/ik/plugins"
	"log"
	"os"
	"path"
	"errors"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

type InputOutputFactoryRegistry struct {
	inputFactories  map[string]ik.InputFactory
	outputFactories map[string]ik.OutputFactory
}

func (registry *InputOutputFactoryRegistry) RegisterInputFactory(factory ik.InputFactory) error {
	_, alreadyExists := registry.inputFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("InputFactory named %s already registered", factory.Name()))
	}
	registry.inputFactories[factory.Name()] = factory
	return nil
}

func (registry *InputOutputFactoryRegistry) LookupInputFactory(name string) ik.InputFactory {
	factory, ok := registry.inputFactories[name]
	if !ok {
		return nil
	}
	return factory
}

func (registry *InputOutputFactoryRegistry) RegisterOutputFactory(factory ik.OutputFactory) error {
	_, alreadyExists := registry.outputFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("OutputFactory named %s already registered", factory.Name()))
	}
	registry.outputFactories[factory.Name()] = factory
	return nil
}

func (registry *InputOutputFactoryRegistry) LookupOutputFactory(name string) ik.OutputFactory {
	factory, ok := registry.outputFactories[name]
	if !ok {
		return nil
	}
	return factory
}

func NewInputOutputFactoryRegistry() *InputOutputFactoryRegistry {
	return &InputOutputFactoryRegistry {
		inputFactories:  make(map[string]ik.InputFactory),
		outputFactories: make(map[string]ik.OutputFactory),
	}
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

	dir, file := path.Split(config_file)
	opener := ik.DefaultOpener(dir)
	config, err := ik.ParseConfig(opener, file)
	if err != nil {
		println(err.Error())
		return
	}

	registry := NewInputOutputFactoryRegistry()

	for _, _plugin := range plugins.GetPlugins() {
		switch plugin := _plugin.(type) {
		case ik.InputFactory:
			registry.RegisterInputFactory(plugin)
		case ik.OutputFactory:
			registry.RegisterOutputFactory(plugin)
		}
	}

	scorekeeper := ik.NewScoreKeeper()
	defer scorekeeper.Dispose()
	router := ik.NewFluentRouter()
	engine := ik.NewEngine(logger, scorekeeper, router)
	defer engine.Dispose()

	err = ik.NewFluentConfigurer(logger, registry, registry, router).Configure(engine, config)
	if err != nil {
		println(err.Error())
		return
	}
	engine.Start()
}

// vim: sts=4 sw=4 ts=4 noet
