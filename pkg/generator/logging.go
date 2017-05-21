package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
)

type loggingGenerator struct {
	*fileGenerator
}

func newLoggingGenerator(source service.Source) *loggingGenerator {
	g := &loggingGenerator{newFileGenerator(source)}
	g.generator = g.generate
	return g
}

func (g *loggingGenerator) generate(file *jen.File) error {
	if err := g.generateStruct(file); err != nil {
		return err
	}

	if err := g.generateConstructor(file); err != nil {
		return err
	}

	if err := g.generateMethods(file); err != nil {
		return err
	}

	return nil
}

func (g *loggingGenerator) generateStruct(file *jen.File) error {
	file.Type().Id("serviceLoggingMiddleware").Struct(
		jen.Id("logger").Qual("github.com/go-kit/kit/log", "Logger"),
		jen.Id("next").Id("Service"),
	)

	return nil
}

func (g *loggingGenerator) generateConstructor(file *jen.File) error {
	file.Comment("ServiceLoggingMiddleware returns a service middleware that logs the")
	file.Comment("parameters and result of each method invocation.")

	file.Func().Id("ServiceLoggingMiddleware").
		Params(jen.Id("logger").Qual("github.com/go-kit/kit/log", "Logger")).
		Id("Middleware").
		Block(
			jen.Return(
				jen.Func().
					Params(jen.Id("next").Id("Service")).
					Id("Service").
					Block(
						jen.Return(
							jen.Id("serviceLoggingMiddleware").
								Values(jen.Dict{
									jen.Id("logger"): jen.Id("logger"),
									jen.Id("next"):   jen.Id("next"),
								})))))

	return nil
}

func (g *loggingGenerator) generateMethods(file *jen.File) error {
	service := g.source.Service()

	for _, method := range service.Methods() {
		file.Commentf("%s implements Service", method.Name())

		file.Func().
			Params(jen.Id("mw").Id("serviceLoggingMiddleware")).
			Id(method.Name()).
			ParamsFunc(func(g *jen.Group) {
				for _, param := range method.Params() {
					g.Id(param.Name()).Id(param.Type())
				}
			}).
			ParamsFunc(func(g *jen.Group) {
				for _, result := range method.Results() {
					g.Id(result.Name()).Id(result.Type())
				}
			}).
			BlockFunc(func(g *jen.Group) {
				g.Defer().Func().
					Params(jen.Id("begin").Id("time.Time")).
					BlockFunc(func(g *jen.Group) {
						g.Id("mw.logger.Log").CallFunc(func(g *jen.Group) {
							g.Lit("method")
							g.Lit(method.Name())

							g.Lit("took")
							g.Id("time.Since").Call(jen.Id("begin"))

							for _, result := range method.Results() {
								if result.Type() == "error" {
									g.Lit("err")
									g.Id(result.Name())
								}
							}
						})
					}).
					Call(jen.Id("time.Now").Call()).Line().
					Return().Id(fmt.Sprintf("mw.next.%s", method.Name())).CallFunc(func(g *jen.Group) {
					for _, param := range method.Params() {
						g.Id(param.Name())
					}
				})
			})
	}

	return nil
}
