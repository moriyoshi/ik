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

type PluginInstance interface {
	Spawnee
	Factory() Plugin
}

type Input interface {
	PluginInstance
	Port() Port
}

type Output interface {
	PluginInstance
	Port
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
	PlainText(PluginInstance) (string, error)
	Markup(PluginInstance) (Markup, error)
}

type Disposable interface {
	Dispose()
}

type Plugin interface {
	Name() string
	BindScorekeeper(*Scorekeeper)
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
	Launch(PluginInstance) error
	SpawneeStatuses() ([]SpawneeStatus, error)
	PluginInstances() []PluginInstance
}

type InputFactory interface {
	Plugin
	New(engine Engine, config *ConfigElement) (Input, error)
}

type InputFactoryRegistry interface {
	RegisterInputFactory(factory InputFactory) error
	LookupInputFactory(name string) InputFactory
}

type OutputFactory interface {
	Plugin
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
	PluginInstance
}

type ScoreboardFactory interface {
	Plugin
	New(engine Engine, pluginRegistry PluginRegistry, config *ConfigElement) (Scoreboard, error)
}
