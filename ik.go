package ik

import "log"

type FluentRecord struct {
	Tag       string
	Timestamp uint64
	Data      map[string]interface{}
}

type Port interface {
	Emit(record []FluentRecord) error
}

type Spawnee interface {
	Run() error
	Shutdown() error
}

type Input interface {
	Spawnee
	Factory() InputFactory
	Port() Port
}

type Output interface {
	Port
	Spawnee
	Factory() OutputFactory
}

type MarkupAttributes int

const (
	Embolden   = 0x10000
	Underlined = 0x20000
)

type MarkupChunk struct {
	Attrs MarkupAttributes
	Text  string
}

type Markup struct {
	Chunks []MarkupChunk
}

type ScoreValueFetcher interface {
	PlainText() (string, error)
	Markup() (Markup, error)
}

type Disposable interface {
	Dispose()
}

type Plugin interface {
	Name() string
}

type ScorekeeperTopic struct {
	Plugin Plugin
	Name string
	DisplayName string
	Description string
	Fetcher ScoreValueFetcher
}

type Engine interface {
	Disposable
	Logger() *log.Logger
	Scorekeeper() *Scorekeeper
	DefaultPort() Port
	Spawn(Spawnee) error
	SpawneeStatuses() ([]SpawneeStatus, error)
}

type InputFactory interface {
	Name() string
	New(engine Engine, config *ConfigElement) (Input, error)
}

type InputFactoryRegistry interface {
	RegisterInputFactory(factory InputFactory) error
	LookupInputFactory(name string) InputFactory
}

type OutputFactory interface {
	Name() string
	New(engine Engine, config *ConfigElement) (Output, error)
}

type OutputFactoryRegistry interface {
	RegisterOutputFactory(factory OutputFactory) error
	LookupOutputFactory(name string) OutputFactory
}

type PluginRegistry interface {
	Plugins() []Plugin
}

type Scoreboard interface {
	Spawnee
	Factory() ScoreboardFactory
}

type ScoreboardFactory interface {
	Name() string
	New(engine Engine, pluginRegistry PluginRegistry, config *ConfigElement) (Scoreboard, error)
}
