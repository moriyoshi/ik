package ik

import (
	"log"
)

type engineImpl struct {
	logger          *log.Logger
	scorekeeper     *Scorekeeper
	defaultPort     Port
	spawner         *Spawner
	pluginInstances []PluginInstance
}

func (engine *engineImpl) Logger() *log.Logger {
	return engine.logger
}

func (engine *engineImpl) Scorekeeper() *Scorekeeper {
	return engine.scorekeeper
}

func (engine *engineImpl) SpawneeStatuses() ([]SpawneeStatus, error) {
	return engine.spawner.GetSpawneeStatuses()
}

func (engine *engineImpl) DefaultPort() Port {
	return engine.defaultPort
}

func (engine *engineImpl) Dispose() {
	spawnees, err := engine.spawner.GetRunningSpawnees()
	if err != nil {
		engine.logger.Print(err.Error())
	} else {
		for _, spawnee := range spawnees {
			engine.spawner.Kill(spawnee)
		}
		engine.spawner.PollMultiple(spawnees)
	}
}

func (engine *engineImpl) Spawn(spawnee Spawnee) error {
	return engine.spawner.Spawn(spawnee)
}

func (engine *engineImpl) Launch(pluginInstance PluginInstance) error {
	var err error
	spawnee, ok := pluginInstance.(Spawnee)
	if ok {
		err = engine.Spawn(spawnee)
		if err != nil {
			return err
		}
	}
	engine.pluginInstances = append(engine.pluginInstances, pluginInstance)
	return nil
}

func (engine *engineImpl) PluginInstances() []PluginInstance {
	retval := make([]PluginInstance, len(engine.pluginInstances))
	copy(retval, engine.pluginInstances)
	return retval
}

func (engine *engineImpl) Start() error {
	spawnees, err := engine.spawner.GetRunningSpawnees()
	if err != nil {
		return err
	}
	return engine.spawner.PollMultiple(spawnees)
}

func NewEngine(logger *log.Logger, scorekeeper *Scorekeeper, defaultPort Port) *engineImpl {
	engine := &engineImpl{
		logger:          logger,
		scorekeeper:     scorekeeper,
		defaultPort:     defaultPort,
		spawner:         NewSpawner(),
		pluginInstances: make([]PluginInstance, 0),
	}
	return engine
}
