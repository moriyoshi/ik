package ik

import "sync"

type descriptorListHead struct {
	next *spawneeDescriptor
	prev *spawneeDescriptor
}

type descriptorList struct {
	count int
	first *spawneeDescriptor
	last *spawneeDescriptor
}

type spawneeDescriptor struct {
	head_alive descriptorListHead
	head_dead descriptorListHead
	spawnee Spawnee
	exitStatus error
	panic interface{}
	shutdownRequested bool
	mtx sync.Mutex
	cond *sync.Cond
}

type dispatchReturnValue struct {
	b bool
	s []Spawnee
	i interface{}
	e error
}

type dispatch struct {
	fn func(Spawnee, chan dispatchReturnValue)
	spawnee Spawnee
	retval chan dispatchReturnValue
}

const (
	Spawned = 1
	Stopped = 2
)

type ContinueType struct {}

func (_ *ContinueType) Error() string { return "" }

var Continue = &ContinueType {}

type NotFoundType struct {}

func (_ *NotFoundType) Error() string { return "not found" }

var NotFound = &NotFoundType {}

type spawnerEvent struct {
	t int
	spawnee Spawnee
}

type Spawner struct {
	alives descriptorList
	deads descriptorList
	m map[Spawnee]*spawneeDescriptor
	c chan dispatch
	mtx sync.Mutex
	cond *sync.Cond
	lastEvent *spawnerEvent
}

func newDescriptor(spawnee Spawnee) *spawneeDescriptor {
	retval := &spawneeDescriptor {
		head_alive: descriptorListHead { nil, nil },
		head_dead: descriptorListHead { nil, nil },
		spawnee: spawnee,
		exitStatus: Continue,
		panic: nil,
		shutdownRequested: false,
		mtx: sync.Mutex {},
		cond: nil,
	}
	retval.cond = sync.NewCond(&retval.mtx)
	return retval
}

func (spawner *Spawner) spawn(spawnee Spawnee, retval chan dispatchReturnValue) {
	go func() {
		descriptor := newDescriptor(spawnee)
		func() {
			spawner.mtx.Lock()
			defer spawner.mtx.Unlock()
			if spawner.alives.last != nil {
				spawner.alives.last.head_alive.next = descriptor
				descriptor.head_alive.prev = spawner.alives.last
			}
			if spawner.alives.first == nil {
				spawner.alives.first = descriptor
			}
			spawner.alives.last = descriptor
			spawner.alives.count += 1
			spawner.m[spawnee] = descriptor
			// notify the event
			spawner.lastEvent = &spawnerEvent {
				t: Spawned,
				spawnee: spawnee,
			}
			spawner.cond.Broadcast()
			retval <- dispatchReturnValue { true, nil, nil, nil }
		}()
		var exitStatus error = nil
		var panic_ interface{} = nil
		func() {
			defer func() {
				r := recover()
				if r != nil {
					exitStatus = nil
					panic_ = r
				}
			}()
			exitStatus = Continue
			for exitStatus == Continue {
				exitStatus = descriptor.spawnee.Run()
			}
		}()
		func() {
			spawner.mtx.Lock()
			defer spawner.mtx.Unlock()
			descriptor.exitStatus = exitStatus
			descriptor.panic = panic_
			// remove from alive list
			if descriptor.head_alive.prev != nil {
				descriptor.head_alive.prev.head_alive.next = descriptor.head_alive.next
			} else {
				spawner.alives.first = descriptor.head_alive.next
			}
			if descriptor.head_alive.next != nil {
				descriptor.head_alive.next.head_alive.prev = descriptor.head_alive.prev
			} else {
				spawner.alives.last = descriptor.head_alive.prev
			}
			spawner.alives.count -= 1
			// append to dead list
			if spawner.deads.last != nil {
				spawner.deads.last.head_dead.next = descriptor
				descriptor.head_dead.prev = spawner.deads.last
			}
			if spawner.deads.first == nil {
				spawner.deads.first = descriptor
			}
			spawner.deads.last = descriptor
			spawner.deads.count += 1

			// notify the event
			spawner.lastEvent = &spawnerEvent {
				t: Stopped,
				spawnee: spawnee,
			}
			spawner.cond.Broadcast()
			descriptor.cond.Broadcast()
		}()
	}()
}

func (spawner *Spawner) kill(spawnee Spawnee, retval chan dispatchReturnValue) {
	spawner.mtx.Lock()
	defer spawner.mtx.Unlock()
	descriptor, ok := spawner.m[spawnee]
	if ok && descriptor.exitStatus != Continue {
		descriptor.shutdownRequested = true
		err := spawnee.Shutdown()
		retval <- dispatchReturnValue { true, nil, nil, err }
		return
	}
	retval <- dispatchReturnValue { false, nil, nil, nil }
}

func (spawner *Spawner) getStatus(spawnee Spawnee, retval chan dispatchReturnValue) {
	spawner.mtx.Lock()
	defer spawner.mtx.Unlock()
	descriptor, ok := spawner.m[spawnee]
	if ok {
		retval <- dispatchReturnValue { false, nil, descriptor.panic, descriptor.exitStatus }
	} else {
		retval <- dispatchReturnValue { false, nil, nil, nil }
	}
}

func (spawner *Spawner) getRunningSpawnees(_ Spawnee, retval chan dispatchReturnValue) {
	spawner.mtx.Lock()
	defer spawner.mtx.Unlock()
	spawnees := make([]Spawnee, spawner.alives.count)
	descriptor := spawner.alives.first
	i := 0
	for descriptor != nil {
		spawnees[i] = descriptor.spawnee
		descriptor = descriptor.head_alive.next
		i += 1
	}
	retval <- dispatchReturnValue { false, spawnees, nil, nil }
}

func (spawner *Spawner) getStoppedSpawnees(_ Spawnee, retval chan dispatchReturnValue) {
	spawner.mtx.Lock()
	defer spawner.mtx.Unlock()
	spawnees := make([]Spawnee, spawner.deads.count)
	descriptor := spawner.deads.first
	i := 0
	for descriptor != nil {
		spawnees[i] = descriptor.spawnee
		descriptor = descriptor.head_dead.next
		i += 1
	}
	retval <- dispatchReturnValue { false, spawnees, nil, nil }
}

func (spawner *Spawner) Spawn(spawnee Spawnee) error {
	retval := make(chan dispatchReturnValue)
	spawner.c <- dispatch { spawner.spawn, spawnee, retval }
	retval_ := <-retval
	return retval_.e
}

func (spawner *Spawner) Kill(spawnee Spawnee) (bool, error) {
	retval := make(chan dispatchReturnValue)
	spawner.c <- dispatch { spawner.kill, spawnee, retval }
	retval_ := <-retval
	return retval_.b, retval_.e
}

func (spawner *Spawner) GetStatus(spawnee Spawnee) (error, interface{}) {
	retval := make(chan dispatchReturnValue)
	spawner.c <- dispatch { spawner.getStatus, spawnee, retval }
	retval_ := <-retval
	return retval_.e, retval_.i
}

func (spawner *Spawner) GetRunningSpawnees() ([]Spawnee, error) {
	retval := make(chan dispatchReturnValue)
	spawner.c <- dispatch { spawner.getRunningSpawnees, nil, retval }
	retval_ := <-retval
	return retval_.s, retval_.e
}

func (spawner *Spawner) GetStoppedSpawnees() ([]Spawnee, error) {
	retval := make(chan dispatchReturnValue)
	spawner.c <- dispatch { spawner.getStoppedSpawnees, nil, retval }
	retval_ := <-retval
	return retval_.s, retval_.e
}

func (spawner *Spawner) Poll(spawnee Spawnee) error {
	descriptor, ok := spawner.m[spawnee]
	if !ok {
		return NotFound
	}
	if func() bool {
		spawner.mtx.Lock()
		defer spawner.mtx.Unlock()
		if descriptor.exitStatus != Continue {
			return true
		}
		return false
	}() {
		return nil
	}
	defer descriptor.mtx.Unlock()
	descriptor.mtx.Lock()
	descriptor.cond.Wait()
	return nil
}

func (spawner *Spawner) PollMultiple(spawnees []Spawnee) error {
	spawnees_ := make(map[Spawnee]bool)
	for _, spawnee := range spawnees {
		spawnees_[spawnee] = true
	}
	count := len(spawnees_)
	for count > 0 {
		spawner.mtx.Lock()
		spawner.cond.Wait()
		lastEvent := spawner.lastEvent
		spawner.mtx.Unlock()
		if lastEvent.t == Stopped {
			spawnee := lastEvent.spawnee
			if alive, ok := spawnees_[spawnee]; alive && ok {
				spawnees_[spawnee] = false
				count -= 1
			}
		}
	}
	return nil
}

func NewSpawner() *Spawner {
	c := make(chan dispatch)
	// launch the supervisor
	go func() {
		for {
			select {
			case disp := <-c:
				disp.fn(disp.spawnee, disp.retval)
			}
		}
	}()
	spawner := &Spawner {
		alives: descriptorList { 0, nil, nil },
		deads: descriptorList { 0, nil, nil },
		m: make(map[Spawnee]*spawneeDescriptor),
		c: c,
		mtx: sync.Mutex {},
		cond: nil,
		lastEvent: nil,
	}
	spawner.cond = sync.NewCond(&spawner.mtx)
	return spawner
}
