package ik

import (
    "log"
)

type topicEntry struct {
    plugin Plugin
    name string
}

type scorekeeperImpl struct {
    logger *log.Logger
    engine Engine
    topics map[Plugin]map[string]topicEntry
}

func (sk *scorekeeperImpl) AddTopic(plugin Plugin, name string) {
    sk.logger.Printf("AddTopic: plugin=%s, name=%s", plugin.Name(), name)
    entries, ok := sk.topics[plugin]
    if !ok {
        entries = make(map[string]topicEntry)
        sk.topics[plugin] = entries
    }
    entries[name] = topicEntry { plugin, name }
}

func (sk *scorekeeperImpl) emitInner(entry topicEntry, data ScoreValue) {
    // TODO
    sk.logger.Printf("scorekeeper: plugin=%s, name=%s, data=%s", entry.plugin.Name(), entry.name, data.AsPlainText())
}

func (sk *scorekeeperImpl) Emit(plugin Plugin, name string, data ScoreValue) {
    var ok bool
    var entries map[string]topicEntry
    var entry topicEntry
    entries, ok = sk.topics[plugin]
    if ok { entry, ok = entries[name] }
    if !ok {
        sk.logger.Printf("unknown topic: plugin=%s, name=%s", plugin.Name(), name)
        return
    }
    sk.emitInner(entry, data)
}

func (sk *scorekeeperImpl) Bind(engine Engine) {
    sk.engine = engine
    sk.logger = engine.Logger()
}

func NewScoreKeeper() ScoreKeeper {
    return &scorekeeperImpl {
        logger: nil,
        engine: nil,
        topics: make(map[Plugin]map[string]topicEntry),
    }
}
