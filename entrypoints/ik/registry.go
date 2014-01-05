package main

import (
	"fmt"
	"errors"
	"github.com/moriyoshi/ik"
)

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
