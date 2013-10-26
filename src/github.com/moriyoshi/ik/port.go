package ik

import "log"

type demoPort struct {
    logger *log.Logger
}

func (port *demoPort) Emit(record FluentRecord) error {
    port.logger.Printf("tag=%s, timestamp=%d, data=%s\n", record.Tag, record.Timestamp, record.Data)
    return nil
}

func newDemoPort(logger *log.Logger) *demoPort {
    return &demoPort { logger: logger }
}
