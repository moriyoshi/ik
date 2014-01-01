package ik

import (
	"log"
)

type engineImpl struct {
	logger          *log.Logger
	scorekeeper     ScoreKeeper
	defaultPort     Port
	spawner         *Spawner
}

func (engine *engineImpl) Logger() *log.Logger {
	return engine.logger
}

func (engine *engineImpl) ScoreKeeper() ScoreKeeper {
	return engine.scorekeeper
}

func (engine *engineImpl) DefaultPort() Port {
	return engine.defaultPort
}

func (engine *engineImpl) Dispose() {
	spawnees, err := engine.spawner.GetRunningSpawnees()
	if err != nil {
		engine.logger.Fatal(err.Error())
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

func (engine *engineImpl) Start() error {
	spawnees, err := engine.spawner.GetRunningSpawnees()
	if err != nil {
		return err
	}
	return engine.spawner.PollMultiple(spawnees)
}

func NewEngine(logger *log.Logger, scorekeeper ScoreKeeper, defaultPort Port) *engineImpl {
	engine := &engineImpl{
		logger:          logger,
		scorekeeper:     scorekeeper,
		defaultPort:     defaultPort,
		spawner:         NewSpawner(),
	}
	scorekeeper.Bind(engine)
	return engine
}
