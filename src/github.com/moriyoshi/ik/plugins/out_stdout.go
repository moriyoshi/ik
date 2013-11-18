package plugins

import (
	"fmt"
	"github.com/moriyoshi/ik"
	"log"
	"os"
)

type StdoutOutput struct {
	factory *StdoutOutputFactory
	logger  *log.Logger
}

func (output *StdoutOutput) Emit(record []ik.FluentRecord) error {
	for _, record := range record {
		fmt.Fprintf(os.Stdout, "%d %s: %s\n", record.Timestamp, record.Tag, record.Data)
	}
	return nil
}

func (output *StdoutOutput) Factory() ik.OutputFactory {
	return output.factory
}

type StdoutOutputFactory struct {
}

func newStdoutOutput(factory *StdoutOutputFactory, logger *log.Logger) (*StdoutOutput, error) {
	return &StdoutOutput{
		factory: factory,
		logger:  logger,
	}, nil
}

func (factory *StdoutOutputFactory) Name() string {
	return "stdout"
}

func (factory *StdoutOutputFactory) New(engine ik.Engine, attrs map[string]string) (ik.Output, error) {
	return newStdoutOutput(factory, engine.Logger())
}

var singleton2 = StdoutOutputFactory{}

func GetStdoutOutputFactory() *StdoutOutputFactory {
	return &singleton2
}
