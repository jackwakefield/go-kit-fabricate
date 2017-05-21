package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
)

type instrumentingGenerator struct {
	*fileGenerator
}

func newInstrumentingGenerator(source service.Source) *instrumentingGenerator {
	g := &instrumentingGenerator{newFileGenerator(source)}
	g.generator = g.generate
	return g
}

func (g *instrumentingGenerator) generate(file *jen.File) error {
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

func (g *instrumentingGenerator) generateStruct(file *jen.File) error {
	file.Type().Id("serviceInstrumentingMiddleware").StructFunc(func(g *jen.Group) {
		g.Id("requestCount").Id("metrics.Counter")
		g.Id("errorCount").Id("metrics.Counter")
		g.Id("requestLatency").Id("metrics.Histogram")
		g.Id("next").Id("Service")
	})

	return nil
}

func (g *instrumentingGenerator) generateConstructor(file *jen.File) error {
	file.Func().Id("ServiceIntrumentingMiddleware").
		Params(
			jen.Id("requestCount"),
			jen.Id("errorCount").Id("metrics.Counter"),
			jen.Id("requestLatency").Id("metrics.Histogram")).
		Id("Middleware").
		Block(
			jen.Return(
				jen.Func().
					Params(jen.Id("next").Id("Service")).
					Id("Service").
					Block(
						jen.Return(
							jen.Id("serviceInstrumentingMiddleware").
								Values(jen.Dict{
									jen.Id("requestCount"):   jen.Id("requestCount"),
									jen.Id("errorCount"):     jen.Id("errorCount"),
									jen.Id("requestLatency"): jen.Id("requestLatency"),
									jen.Id("next"):           jen.Id("next"),
								})))))

	return nil
}

func (g *instrumentingGenerator) generateMethods(file *jen.File) error {
	service := g.source.Service()

	for _, method := range service.Methods() {
		file.Commentf("%s implements Service", method.Name())

		file.Func().
			Params(jen.Id("mw").Id("serviceInstrumentingMiddleware")).
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
						g.Id("mw.requestCount.With").Call(jen.Lit("method"), jen.Lit(method.Name())).
							Dot("Add").Call(jen.Lit(1))

						g.Id("mw.requestLatency.With").Call(jen.Lit("method"), jen.Lit(method.Name())).
							Dot("Observe").Call(jen.Id("time.Since").Call(jen.Id("begin")).Dot("Seconds").Call())

						if errorResult := method.ErrorResult(); errorResult != nil {
							g.If(jen.Id(errorResult.Name()).Op("!=").Nil()).
								Block(jen.Id("mw.errorCount.With").Call(jen.Lit("method"), jen.Lit(method.Name())).
									Dot("Add").Call(jen.Lit(1)))
						}
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
