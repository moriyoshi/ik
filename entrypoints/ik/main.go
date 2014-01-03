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

type MultiFactoryRegistry struct {
	scorekeeper         *ik.Scorekeeper
	inputFactories      map[string]ik.InputFactory
	outputFactories     map[string]ik.OutputFactory
	scoreboardFactories map[string]ik.ScoreboardFactory
	plugins             []ik.Plugin
}

func (registry *MultiFactoryRegistry) RegisterInputFactory(factory ik.InputFactory) error {
	_, alreadyExists := registry.inputFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("InputFactory named %s already registered", factory.Name()))
	}
	registry.inputFactories[factory.Name()] = factory
	registry.plugins = append(registry.plugins, factory)
	factory.BindScorekeeper(registry.scorekeeper)
	return nil
}

func (registry *MultiFactoryRegistry) LookupInputFactory(name string) ik.InputFactory {
	factory, ok := registry.inputFactories[name]
	if !ok {
		return nil
	}
	return factory
}

func (registry *MultiFactoryRegistry) RegisterOutputFactory(factory ik.OutputFactory) error {
	_, alreadyExists := registry.outputFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("OutputFactory named %s already registered", factory.Name()))
	}
	registry.outputFactories[factory.Name()] = factory
	registry.plugins = append(registry.plugins, factory)
	factory.BindScorekeeper(registry.scorekeeper)
	return nil
}

func (registry *MultiFactoryRegistry) LookupOutputFactory(name string) ik.OutputFactory {
	factory, ok := registry.outputFactories[name]
	if !ok {
		return nil
	}
	return factory
}

func (registry *MultiFactoryRegistry) RegisterScoreboardFactory(factory ik.ScoreboardFactory) error {
	_, alreadyExists := registry.scoreboardFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("ScoreboardFactory named %s already registered", factory.Name()))
	}
	registry.scoreboardFactories[factory.Name()] = factory
	registry.plugins = append(registry.plugins, factory)
	factory.BindScorekeeper(registry.scorekeeper)
	return nil
}

func (registry *MultiFactoryRegistry) LookupScoreboardFactory(name string) ik.ScoreboardFactory {
	factory, ok := registry.scoreboardFactories[name]
	if !ok {
		return nil
	}
	return factory
}

func (registry *MultiFactoryRegistry) Plugins() []ik.Plugin {
	retval := make([]ik.Plugin, len(registry.plugins))
	copy(retval, registry.plugins)
	return retval
}

func NewMultiFactoryRegistry(scorekeeper *ik.Scorekeeper) *MultiFactoryRegistry {
	return &MultiFactoryRegistry {
		scorekeeper: scorekeeper,
		inputFactories:  make(map[string]ik.InputFactory),
		outputFactories: make(map[string]ik.OutputFactory),
		scoreboardFactories: make(map[string]ik.ScoreboardFactory),
	}
}

func configureScoreboards(logger *log.Logger, registry *MultiFactoryRegistry, engine ik.Engine, config *ik.Config) error {
       for _, v := range config.Root.Elems {
	       switch v.Name {
		case "scoreboard":
			type_ := v.Attrs["type"]
			scoreboardFactory := registry.LookupScoreboardFactory(type_)
			if scoreboardFactory == nil {
				return errors.New("Could not find scoreboard factory: " + type_)
			}
			scoreboard, err := scoreboardFactory.New(engine, registry, v)
			if err != nil {
				return err
			}
			err = engine.Launch(scoreboard)
			if err != nil {
				return err
			}
			logger.Printf("Scoreboard plugin loaded: %s", scoreboardFactory.Name())
		}
	}
	return nil
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

	scorekeeper := ik.NewScorekeeper(logger)

	registry := NewMultiFactoryRegistry(scorekeeper)

	for _, _plugin := range plugins.GetPlugins() {
		switch plugin := _plugin.(type) {
		case ik.InputFactory:
			registry.RegisterInputFactory(plugin)
		case ik.OutputFactory:
			registry.RegisterOutputFactory(plugin)
		}
	}

	registry.RegisterScoreboardFactory(&HTMLHTTPScoreboardFactory {})

	router := ik.NewFluentRouter()
	engine := ik.NewEngine(logger, scorekeeper, router)
	defer engine.Dispose()

	err = ik.NewFluentConfigurer(logger, registry, registry, router).Configure(engine, config)
	if err != nil {
		println(err.Error())
		return
	}
	err = configureScoreboards(logger, registry, engine, config)
	if err != nil {
		println(err.Error())
		return
	}
	engine.Start()
}

// vim: sts=4 sw=4 ts=4 noet
