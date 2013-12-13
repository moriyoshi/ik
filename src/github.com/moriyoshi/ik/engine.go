package ik

import (
	"errors"
	"fmt"
	"log"
)

type engineImpl struct {
	logger          *log.Logger
	scorekeeper     ScoreKeeper
	inputFactories  map[string]InputFactory
	outputFactories map[string]OutputFactory
	defaultPort     Port
}

func (engine *engineImpl) Logger() *log.Logger {
	return engine.logger
}

func (engine *engineImpl) ScoreKeeper() ScoreKeeper {
	return engine.scorekeeper
}

func (engine *engineImpl) RegisterInputFactory(factory InputFactory) error {
	_, alreadyExists := engine.inputFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("InputFactory named %s already registered", factory.Name()))
	}
	engine.inputFactories[factory.Name()] = factory
	return nil
}

func (engine *engineImpl) LookupInputFactory(name string) InputFactory {
	factory, ok := engine.inputFactories[name]
	if !ok {
		engine.logger.Printf("InputFactory named %s does not exist", name)
		return nil
	}
	return factory
}

func (engine *engineImpl) RegisterOutputFactory(factory OutputFactory) error {
	_, alreadyExists := engine.outputFactories[factory.Name()]
	if alreadyExists {
		return errors.New(fmt.Sprintf("OutputFactory named %s already registered", factory.Name()))
	}
	engine.outputFactories[factory.Name()] = factory
	return nil
}

func (engine *engineImpl) LookupOutputFactory(name string) OutputFactory {
	factory, ok := engine.outputFactories[name]
	if !ok {
		engine.logger.Printf("OutputFactory named %s does not exist", name)
		return nil
	}
	return factory
}

func (engine *engineImpl) DefaultPort() Port {
	return engine.defaultPort
}

func (engine *engineImpl) SetDefaultPort(port Port) {
	engine.defaultPort = port
}

func NewEngine(logger *log.Logger, scorekeeper ScoreKeeper, defaultPort Port) *engineImpl {
	engine := &engineImpl{
		logger:          logger,
		scorekeeper:     scorekeeper,
		inputFactories:  make(map[string]InputFactory),
		outputFactories: make(map[string]OutputFactory),
		defaultPort:     defaultPort,
	}
	scorekeeper.Bind(engine)
	return engine
}
