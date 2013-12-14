package plugins

import (
	"github.com/moriyoshi/ik"
	"io"
	"bufio"
	"strconv"
	"compress/gzip"
	"os"
	"fmt"
	"log"
	"time"
	"errors"
	"encoding/json"
	strftime "github.com/jehiah/go-strftime"
)

type FlushableWriter interface {
	io.Writer
	Flush() error
}

type FileOutput struct {
	factory *FileOutputFactory
	logger  *log.Logger
	out FlushableWriter
	closer func() error
	timeFormat string
	location *time.Location
	c chan []ik.FluentRecord
	cancel chan bool
}

type FileOutputFactory struct {
}

const (
	compressionNone = 0
	compressionGzip = 1
)

func (output *FileOutput) formatTime(timestamp uint64) string {
	timestamp_ := time.Unix(int64(timestamp), 0)
	if output.timeFormat == "" {
		return timestamp_.Format(time.RFC3339)
	} else {
		return strftime.Format(output.timeFormat, timestamp_)
	}
}

func (output *FileOutput) formatData(data map[string]interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil // XXX: byte => rune
}

func (output *FileOutput) Emit(records []ik.FluentRecord) error {
	output.c <- records
	return nil
}

func (output *FileOutput) Factory() ik.OutputFactory {
	return output.factory
}

func (output *FileOutput) Run() error {
	select {
	case <- output.cancel:
		return nil
	case records := <-output.c:
		for _, record := range records {
			formattedData, err := output.formatData(record.Data)
			if err != nil {
				return err
			}
			fmt.Fprintf(
				output.out,
				"%s\t%s\t%s\n",
				output.formatTime(record.Timestamp),
				record.Tag,
				formattedData,
			)
		}
		err := output.out.Flush()
		if err != nil {
			return err
		}
	}
	return ik.Continue
}

func (output *FileOutput) Shutdown() error {
	output.cancel <- true
	return output.closer()
}

func newFileOutput(factory *FileOutputFactory, logger *log.Logger, path string, timeFormat string, compressionFormat int, symlinkPath string, permission os.FileMode) (*FileOutput, error) {
	var out FlushableWriter
	var closer func() error
	fout, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE | os.O_APPEND, permission)
	if err != nil {
		logger.Fatal("Failed to open " + path)
		return nil, err
	}
	if compressionFormat == compressionGzip {
		gzipout := gzip.NewWriter(fout)
		out = gzipout
		closer = func() error {
			err1 := gzipout.Close()
			err2 := fout.Close()
			if err2 != nil {
				if err1 != nil {
					logger.Fatal("ignored error: " + err1.Error())
				}
				return err2
			}
			if err1 != nil {
				return err1
			}
			return nil
		}
	} else {
		out = bufio.NewWriter(fout)
		closer = func() error {
			return fout.Close()
		}
	}
	if symlinkPath != "" {
		os.Symlink(symlinkPath, path)
	}
	return &FileOutput {
		factory: factory,
		logger:  logger,
		out: out,
		closer: closer,
		timeFormat: timeFormat,
		location: time.UTC,
		c: make(chan []ik.FluentRecord, 100 /* FIXME */),
		cancel: make(chan bool),
	}, nil
}

func (factory *FileOutputFactory) Name() string {
	return "file"
}

func (factory *FileOutputFactory) New(engine ik.Engine, attrs map[string]string) (ik.Output, error) {
	path := ""
	timeFormat := ""
	compressionFormat := compressionNone
	symlinkPath := ""
	permission := 0666
	path, _ = attrs["path"]
	timeFormat, _ = attrs["time_format"]
	compressionFormatStr, ok := attrs["compress"]
	if ok {
		if compressionFormatStr == "gz" || compressionFormatStr == "gzip" {
			compressionFormat = compressionGzip
		} else {
			return nil, errors.New("unknown compression format: " + compressionFormatStr)
		}
	}
	symlinkPath, _ = attrs["symlink_path"]
	permissionStr, ok := attrs["permission"]
	if ok {
		var err error
		permission, err = strconv.Atoi(permissionStr)
		if err != nil {
			return nil, err
		}
	}
	return newFileOutput(
		factory,
		engine.Logger(),
		path,
		timeFormat,
		compressionFormat,
		symlinkPath,
		os.FileMode(permission),
	)
}

var _ = AddPlugin(&FileOutputFactory{})
