package plugins

import (
	"bytes"
	"log"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/moriyoshi/ik"
	"github.com/ugorji/go/codec"
)

type ForwardOutput struct {
	factory       *ForwardOutputFactory
	logger        *log.Logger
	codec         *codec.MsgpackHandle
	bind          string
	enc           *codec.Encoder
	buffer        bytes.Buffer
	emitCh        chan []ik.FluentRecordSet
	shutdown      chan (chan error)
	flushInterval int
	flushWg       sync.WaitGroup
}

func (output *ForwardOutput) encodeEntry(tag string, record ik.TinyFluentRecord) error {
	v := []interface{}{tag, record.Timestamp, record.Data}
	if output.enc == nil {
		output.enc = codec.NewEncoder(&output.buffer, output.codec)
	}
	err := output.enc.Encode(v)
	if err != nil {
		return err
	}
	return err
}

func (output *ForwardOutput) encodeRecordSet(recordSet ik.FluentRecordSet) error {
	v := []interface{}{recordSet.Tag, recordSet.Records}
	if output.enc == nil {
		output.enc = codec.NewEncoder(&output.buffer, output.codec)
	}
	err := output.enc.Encode(v)
	if err != nil {
		return err
	}
	return err
}

func (output *ForwardOutput) flush() error {
	buffer := output.buffer
	output.buffer = bytes.Buffer{}

	output.flushWg.Add(1)
	go func() { // TODO: static goroutine for flushing.
		defer output.flushWg.Done()
		conn, err := net.Dial("tcp", output.bind)
		if err != nil {
			output.logger.Printf("%#v", err)
			return
		}
		defer conn.Close()

		if n, err := buffer.WriteTo(conn); err != nil {
			output.logger.Printf("Write failed. size: %d, buf size: %d, error: %#v", n, output.buffer.Len(), err.Error())
		}
	}()
	return nil
}

func (output *ForwardOutput) Emit(recordSet []ik.FluentRecordSet) error {
	output.emitCh <- recordSet
	return nil
}

func (output *ForwardOutput) emit(recordSet []ik.FluentRecordSet) error {
	for _, recordSet := range recordSet {
		err := output.encodeRecordSet(recordSet)
		if err != nil {
			output.logger.Printf("%#v", err)
			return err
		}
	}
	return nil
}

func (output *ForwardOutput) Factory() ik.Plugin {
	return output.factory
}

func (output *ForwardOutput) Run() error {
	time.Sleep(time.Second)
	return ik.Continue
}

func (output *ForwardOutput) mainLoop() {
	ticker := time.NewTicker(time.Duration(output.flushInterval) * time.Second)
	for {
		select {
		case rs := <-output.emitCh:
			if err := output.emit(rs); err != nil {
				output.logger.Printf("%#v", err)
			}
		case <-ticker.C:
			output.flush()
		case finish := <-output.shutdown:
			close(output.emitCh)
			output.flush()
			output.flushWg.Wait()
			finish <- nil
			return
		}
	}
}

func (output *ForwardOutput) Shutdown() error {
	finish := make(chan error)
	output.shutdown <- finish
	return <-finish
}

type ForwardOutputFactory struct {
}

func newForwardOutput(factory *ForwardOutputFactory, logger *log.Logger, bind string, flushInterval int) *ForwardOutput {
	_codec := codec.MsgpackHandle{}
	_codec.MapType = reflect.TypeOf(map[string]interface{}(nil))
	_codec.RawToString = false
	_codec.StructToArray = true
	return &ForwardOutput{
		factory:       factory,
		logger:        logger,
		codec:         &_codec,
		bind:          bind,
		emitCh:        make(chan []ik.FluentRecordSet),
		shutdown:      make(chan chan error),
		flushInterval: flushInterval,
		flushWg:       sync.WaitGroup{},
	}
}

func (factory *ForwardOutputFactory) Name() string {
	return "forward"
}

func (factory *ForwardOutputFactory) New(engine ik.Engine, config *ik.ConfigElement) (ik.Output, error) {
	host, ok := config.Attrs["host"]
	if !ok {
		host = "localhost"
	}
	netPort, ok := config.Attrs["port"]
	if !ok {
		netPort = "24224"
	}
	flush_interval_str, ok := config.Attrs["flush_interval"]
	if !ok {
		flush_interval_str = "60"
	}
	flushInterval, err := strconv.Atoi(flush_interval_str)
	if err != nil {
		engine.Logger().Print(err.Error())
		return nil, err
	}
	bind := host + ":" + netPort
	output := newForwardOutput(factory, engine.Logger(), bind, flushInterval)
	go output.mainLoop()
	return output, nil
}

func (factory *ForwardOutputFactory) BindScorekeeper(scorekeeper *ik.Scorekeeper) {
}

var _ = AddPlugin(&ForwardOutputFactory{})
