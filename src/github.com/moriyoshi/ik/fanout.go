package ik

type Fanout struct {
    ports []Port
}

func (fanout *Fanout) AddPort(port Port) {
    fanout.ports = append(fanout.ports, port)
}

func (fanout *Fanout) Emit(record FluentRecord) error {
    for _, port := range fanout.ports {
        err := port.Emit(record)
        if err != nil {
            panic("MUST DO SOMETHING GOOD") // TODO
        }
    }
    return nil
}
