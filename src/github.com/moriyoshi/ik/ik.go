package ik

import (
	"errors"
	"log"
	"net"
	"reflect"
	//	"encoding/hex"
	"github.com/ugorji/go/codec"
)

type Client struct {
	logger *log.Logger
	conn   net.Conn
	codec  *codec.MsgpackHandle
	enc    *codec.Encoder
	dec    *codec.Decoder
}

type FluentRecord struct {
	tag       string
	timestamp uint64
	data      map[string]interface{}
}

func (c *Client) decodeEntry() (FluentRecord, error) {
	v := []interface{}{nil, nil, nil}
	err := c.dec.Decode(&v)
	if err != nil {
		return FluentRecord{}, err
	}
	tag, ok := v[0].(string)
	if !ok {
		return FluentRecord{}, errors.New("Failed to decode tag field")
	}
	timestamp, ok := v[1].(uint64)
	if !ok {
		return FluentRecord{}, errors.New("Failed to decode timestamp field")
	}
	data, ok := v[2].(map[string]interface{})
	if !ok {
		return FluentRecord{}, errors.New("Failed to decode data field")
	}
	return FluentRecord{tag: tag, timestamp: timestamp, data: data}, nil
}

func (c *Client) handle() {
	entry, err := c.decodeEntry()
	if err != nil {
		c.logger.Fatal(err.Error())
		return
	}
	c.logger.Printf("tag=%s, timestamp=%d, data=%s\n", entry.tag, entry.timestamp, entry.data)
}

func newClient(logger *log.Logger, conn net.Conn, _codec *codec.MsgpackHandle) *Client {
	return &Client{
		logger: logger,
		conn:   conn,
		codec:  _codec,
		enc:    codec.NewEncoder(conn, _codec),
		dec:    codec.NewDecoder(conn, _codec),
	}
}

func Server(logger *log.Logger, bind string) error {
	_codec := codec.MsgpackHandle{}
	_codec.MapType = reflect.TypeOf(map[string]interface{}(nil))
	_codec.RawToString = true
	if bind == "" {
		bind = ":24224"
	}
	ss, err := net.Listen("tcp", bind)
	if err != nil {
		logger.Fatal(err.Error())
		return err
	}
	for {
		cs, err := ss.Accept()
		if err != nil {
			logger.Fatal(err.Error())
			continue
		}
		go newClient(logger, cs, &_codec).handle()

	}
	return nil
}

// vim: sts=4 sw=4 ts=4 noet
