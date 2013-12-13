package ik

import (
	"testing"
	"errors"
)

type Foo struct {
	state string
	c chan string
}

func (foo *Foo) Run() error {
	foo.state = "run"
	retval := <-foo.c
	foo.state = "stopped"
	return errors.New(retval)
}

func (foo *Foo) Shutdown() error {
	foo.c <- "ok"
	return nil
}

type Bar struct { c chan bool }

func (bar *Bar) Run() error {
	<-bar.c
	panic("PANIC")
}

func (foo *Bar) Shutdown() error {
	return nil
}

func TestSpawner_Spawn(t *testing.T) {
	spawner := NewSpawner()
	f := &Foo { "", make(chan string) }
	spawner.Spawn(f)
	err, _ := spawner.GetStatus(f)
	if err != Continue { t.Fail() }
	f.c <- "result"
	spawner.Poll(f)
	err, _ = spawner.GetStatus(f)
	if err == Continue { t.Fail() }
	if err.Error() != "result" { t.Fail() }
}

func TestSpawner_Panic(t *testing.T) {
	spawner := NewSpawner()
	f := &Bar { make(chan bool) }
	spawner.Spawn(f)
	err, _ := spawner.GetStatus(f)
	if err != Continue { t.Fail() }
	f.c <- false
	spawner.Poll(f)
	err, panic_ := spawner.GetStatus(f)
	if err == Continue { t.Fail() }
	if panic_ != "PANIC" { t.Fail() }
}
