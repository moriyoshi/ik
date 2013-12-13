package plugins

import (
	"errors"
	"fmt"
	"github.com/moriyoshi/ik"
	"github.com/ugorji/go/codec"
	"io"
	"log"
	"net"
	"reflect"
)

type forwardClient struct {
	input  *ForwardInput
	logger *log.Logger
	conn   net.Conn
	codec  *codec.MsgpackHandle
	enc    *codec.Encoder
	dec    *codec.Decoder
}

type ForwardInput struct {
	factory  *ForwardInputFactory
	port     ik.Port
	logger   *log.Logger
	bind     string
	listener net.Listener
	codec    *codec.MsgpackHandle
	clients  map[net.Conn]*forwardClient
}

type ForwardInputFactory struct {
}

func decodeTinyEntries(tag []byte, entries []interface{}) ([]ik.FluentRecord, error) {
	retval := make([]ik.FluentRecord, len(entries))
	for i, _entry := range entries {
		entry := _entry.([]interface{})
		timestamp, ok := entry[0].(uint64)
		if !ok {
			return nil, errors.New("Failed to decode timestamp field")
		}
		data, ok := entry[1].(map[string]interface{})
		if !ok {
			return nil, errors.New("Failed to decode data field")
		}
		retval[i] = ik.FluentRecord{
			Tag:       string(tag),
			Timestamp: timestamp,
			Data:      data,
		}
	}
	return retval, nil
}

func (c *forwardClient) decodeEntries() ([]ik.FluentRecord, error) {
	v := []interface{}{nil, nil, nil}
	err := c.dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	tag, ok := v[0].([]byte)
	if !ok {
		return nil, errors.New("Failed to decode tag field")
	}

	var retval []ik.FluentRecord
	switch timestamp_or_entries := v[1].(type) {
	case uint64:
		timestamp := timestamp_or_entries
		data, ok := v[2].(map[string]interface{})
		if !ok {
			return nil, errors.New("Failed to decode data field")
		}
		retval = []ik.FluentRecord{
			{
				Tag:       string(tag),
				Timestamp: timestamp,
				Data:      data,
			},
		}
	case float64:
		timestamp := uint64(timestamp_or_entries)
		data, ok := v[2].(map[string]interface{})
		if !ok {
			return nil, errors.New("Failed to decode data field")
		}
		retval = []ik.FluentRecord{
			{
				Tag:       string(tag),
				Timestamp: timestamp,
				Data:      data,
			},
		}
	case []interface{}:
		if !ok {
			return nil, errors.New("Unexpected payload format")
		}
		retval, err = decodeTinyEntries(tag, timestamp_or_entries)
		if err != nil {
			return nil, err
		}
	case []byte:
		entries := make([]interface{}, 0)
		err := codec.NewDecoderBytes(timestamp_or_entries, c.codec).Decode(&entries)
		if err != nil {
			return nil, err
		}
		retval, err = decodeTinyEntries(tag, entries)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(fmt.Sprintf("Unknown type: %t", timestamp_or_entries))
	}
	return retval, nil
}

func (c *forwardClient) handle() {
	for {
		entries, err := c.decodeEntries()
		if err == io.EOF {
			break
		} else if err != nil {
			c.logger.Println(err.Error())
			continue
		}
		c.input.Port().Emit(entries)
	}
	c.conn.Close()
	c.input.markDischarged(c)
}

func newForwardClient(input *ForwardInput, logger *log.Logger, conn net.Conn, _codec *codec.MsgpackHandle) *forwardClient {
	c := &forwardClient{
		input:  input,
		logger: logger,
		conn:   conn,
		codec:  _codec,
		enc:    codec.NewEncoder(conn, _codec),
		dec:    codec.NewDecoder(conn, _codec),
	}
	input.markCharged(c)
	return c
}

func (input *ForwardInput) Factory() ik.InputFactory {
	return input.factory
}

func (input *ForwardInput) Port() ik.Port {
	return input.port
}

func (input *ForwardInput) Run() error {
	conn, err := input.listener.Accept()
	if err != nil {
		input.logger.Fatal(err.Error())
		return err
	}
	go newForwardClient(input, input.logger, conn, input.codec).handle()
	return ik.Continue
}

func (input *ForwardInput) Shutdown() error {
	for conn, _ := range input.clients {
		conn.Close()
	}
	return input.listener.Close()
}

func (input *ForwardInput) markCharged(c *forwardClient) {
	input.clients[c.conn] = c
}

func (input *ForwardInput) markDischarged(c *forwardClient) {
	delete(input.clients, c.conn)
}

func newForwardInput(factory *ForwardInputFactory, logger *log.Logger, bind string, port ik.Port) (*ForwardInput, error) {
	_codec := codec.MsgpackHandle{}
	_codec.MapType = reflect.TypeOf(map[string]interface{}(nil))
	_codec.RawToString = false
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		logger.Fatal(err.Error())
		return nil, err
	}
	return &ForwardInput{
		factory:  factory,
		port:     port,
		logger:   logger,
		bind:     bind,
		listener: listener,
		codec:    &_codec,
		clients:  make(map[net.Conn]*forwardClient),
	}, nil
}

func (factory *ForwardInputFactory) Name() string {
	return "forward"
}

func (factory *ForwardInputFactory) New(engine ik.Engine, attrs map[string]string) (ik.Input, error) {
	listen, ok := attrs["listen"]
	if !ok {
		listen = ""
	}
	netPort, ok := attrs["port"]
	if !ok {
		netPort = "24224"
	}
	bind := listen + ":" + netPort
	return newForwardInput(factory, engine.Logger(), bind, engine.DefaultPort())
}

var singleton = ForwardInputFactory{}

func GetForwardInputFactory() *ForwardInputFactory {
	return &singleton
}
