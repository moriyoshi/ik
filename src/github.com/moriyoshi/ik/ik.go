package ik

import "log"

type FluentRecord struct {
	Tag       string
	Timestamp uint64
	Data      map[string]interface{}
}

type Port interface {
    Emit(record FluentRecord) error
}

type Input interface {
    Factory() InputFactory
    Port() Port
    Start() error
    Shutdown() error
}

type Output interface {
    Port
    Factory() OutputFactory
}

type MarkupAttributes int

const (
    Embolden = 0x10000
    Underlined = 0x20000
)

type MarkupChunk struct {
    Attrs MarkupAttributes
    Text string
}

type Markup struct {
    Chunks []MarkupChunk
}

type ScoreValue interface {
    AsPlainText() string
    AsMarkup() Markup
}

type Plugin interface {
    Name() string
}

type ScoreKeeper interface {
    Bind(engine Engine)
    AddTopic(plugin Plugin, name string)
    Emit(plugin Plugin, name string, data ScoreValue)
}

type Engine interface {
    Logger() *log.Logger
    ScoreKeeper() ScoreKeeper
    DefaultPort() Port
}

type InputFactory interface {
    Name() string
    New(engine Engine, attrs map[string]string) (Input, error)
}

type InputFactoryRegistry interface {
    RegisterInputFactory(factory InputFactory) error
    LookupInputFactory(name string) InputFactory
}

type OutputFactory interface {
    Name() string
    New(engine Engine, attrs map[string]string) (Output, error)
}

type OutputFactoryRegistry interface {
    RegisterOutputFactory(factory OutputFactory) error
    LookupOutputFactory(name string) OutputFactory
}
