package generator

import (
	"path/filepath"

	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
)

type Generator interface {
	GenerateMiddleware() error

	GenerateLogging() error

	GenerateInstrumenting() error

	GenerateEndpoints() error
}

type generator struct {
	source service.Source
	dir    string
}

func NewGenerator(source service.Source, dir string) Generator {
	return &generator{source, dir}
}

func (g *generator) GenerateMiddleware() error {
	path := filepath.Join(g.dir, "middleware_fabricate.go")
	middleware := newMiddlewareGenerator(g.source)
	return middleware.GenerateFile(path)
}

func (g *generator) GenerateLogging() error {
	path := filepath.Join(g.dir, "logging_fabricate.go")
	logging := newLoggingGenerator(g.source)
	return logging.GenerateFile(path)
}

func (g *generator) GenerateInstrumenting() error {
	path := filepath.Join(g.dir, "instrumenting_fabricate.go")
	instrumenting := newInstrumentingGenerator(g.source)
	return instrumenting.GenerateFile(path)
}

func (g *generator) GenerateEndpoints() error {
	path := filepath.Join(g.dir, "endpoints_fabricate.go")
	endpoints := newEndpointsGenerator(g.source)
	return endpoints.GenerateFile(path)
}
