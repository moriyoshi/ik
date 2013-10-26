package plugins

import (
    "io"
	"errors"
	"log"
	"net"
	"reflect"
	//	"encoding/hex"
	"github.com/moriyoshi/ik"
	"github.com/ugorji/go/codec"
)

type streamClient struct {
    input *StreamInput
	logger *log.Logger
	conn   net.Conn
	codec  *codec.MsgpackHandle
	enc    *codec.Encoder
	dec    *codec.Decoder
}

type StreamInput struct {
    factory *StreamInputFactory
    port ik.Port
    logger *log.Logger
    bind string
    listener net.Listener
    codec *codec.MsgpackHandle
    running bool
    clients map[net.Conn]*streamClient
}

type StreamInputFactory struct {
}

func (c *streamClient) decodeEntry() (ik.FluentRecord, error) {
	v := []interface{}{nil, nil, nil}
	err := c.dec.Decode(&v)
	if err != nil {
		return ik.FluentRecord{}, err
	}
	tag, ok := v[0].(string)
	if !ok {
		return ik.FluentRecord{}, errors.New("Failed to decode tag field")
	}
	timestamp, ok := v[1].(uint64)
	if !ok {
		return ik.FluentRecord{}, errors.New("Failed to decode timestamp field")
	}
	data, ok := v[2].(map[string]interface{})
	if !ok {
		return ik.FluentRecord{}, errors.New("Failed to decode data field")
	}
	return ik.FluentRecord{Tag: tag, Timestamp: timestamp, Data: data}, nil
}

func (c *streamClient) handle() {
    for {
        entry, err := c.decodeEntry()
        if err == io.EOF {
            break
        } else if err != nil {
            c.logger.Fatal(err.Error())
            return
        }
        c.input.Port().Emit(entry)
    }
    c.conn.Close()
    c.input.markDischarged(c)
}

func newStreamClient(input *StreamInput, logger *log.Logger, conn net.Conn, _codec *codec.MsgpackHandle) *streamClient {
	c := &streamClient{
        input: input,
		logger: logger,
		conn:   conn,
		codec:  _codec,
		enc:    codec.NewEncoder(conn, _codec),
		dec:    codec.NewDecoder(conn, _codec),
	}
    input.markCharged(c)
    return c
}

func (input *StreamInput) Factory() ik.InputFactory {
    return input.factory
}

func (input *StreamInput) Port() ik.Port {
    return input.port
}

func (input *StreamInput) Start() error {
    input.running = true // XXX: RACE
    go func () {
        for input.running {
            conn, err := input.listener.Accept()
            if err != nil {
                input.logger.Fatal(err.Error())
                continue
            }
            go newStreamClient(input, input.logger, conn, input.codec).handle()
        }
    }()
	return nil
}

func (input *StreamInput) Shutdown() error {
    input.running = false
    for conn, _ := range input.clients {
        conn.Close()
    }
    return input.listener.Close()
}

func (input *StreamInput) markCharged(c *streamClient) {
    input.clients[c.conn] = c
}

func (input *StreamInput) markDischarged(c *streamClient) {
    delete(input.clients, c.conn)
}

func newStreamInput(factory *StreamInputFactory, logger *log.Logger, bind string, port ik.Port) (*StreamInput, error) {
	_codec := codec.MsgpackHandle{}
	_codec.MapType = reflect.TypeOf(map[string]interface{}(nil))
	_codec.RawToString = true
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		logger.Fatal(err.Error())
		return nil, err
	}
    return &StreamInput {
        factory: factory,
        port: port,
        logger: logger,
        bind: bind,
        listener: listener,
        codec: &_codec,
        running: false,
        clients: make(map[net.Conn]*streamClient),
    }, nil
}

func (factory *StreamInputFactory) Name() string {
    return "forward"
}

func (factory *StreamInputFactory) New(engine ik.Engine, attrs map[string]string) (ik.Input, error) {
    listen, ok := attrs["listen"]
    if !ok { listen = "" }
    netPort, ok := attrs["port"]
    if !ok { netPort = "24224" }
    bind := listen + ":" + netPort
    return newStreamInput(factory, engine.Logger(), bind, engine.DefaultPort())
}

var singleton = StreamInputFactory {}

func GetStreamInputFactory() *StreamInputFactory {
    return &singleton
}
