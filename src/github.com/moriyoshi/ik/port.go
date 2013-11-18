package ik

import "log"
import "container/list"

type demoPort struct {
    logger *log.Logger
    outputs *list.List
}

func (port *demoPort) RegisterOutput(output Output) {
    port.outputs.PushBack(output)
}

func (port *demoPort) Emit(record []FluentRecord) error {
    for _, record := range record {
        port.logger.Printf("tag=%s, timestamp=%d, data=%s\n", record.Tag, record.Timestamp, record.Data)
    }
    for output := port.outputs.Front(); output != nil; output = output.Next() {
        o := output.Value.(Output)
        o.Emit(record)
    }
    return nil
}

func newDemoPort(logger *log.Logger) *demoPort {
    return &demoPort { logger: logger, outputs: list.New() }
}
