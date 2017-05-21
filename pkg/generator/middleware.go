package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
)

type middlewareGenerator struct {
	*fileGenerator
}

func newMiddlewareGenerator(source service.Source) *middlewareGenerator {
	g := &middlewareGenerator{newFileGenerator(source)}
	g.generator = g.generate
	return g
}

func (g *middlewareGenerator) generate(file *jen.File) error {
	// define the Middleware function type
	file.Comment("Middleware describes a service (as opposed to endpoint) middleware.")
	file.Type().Id("Middleware").Func().
		Params(jen.Id("Service")).
		Id("Service")

	return nil
}
