package generator

import (
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
	file.Type().Id("serviceInstrumentingMiddleware").Struct(
		jen.Id("requestCount").Qual("github.com/go-kit/kit/metrics", "Counter"),
		jen.Id("errorCount").Qual("github.com/go-kit/kit/metrics", "Counter"),
		jen.Id("requestLatency").Qual("github.com/go-kit/kit/metrics", "Histogram"),
		jen.Id("next").Id("Service"),
	)

	return nil
}

func (g *instrumentingGenerator) generateConstructor(file *jen.File) error {
	file.Func().Id("ServiceIntrumentingMiddleware").
		Params(
			jen.Id("requestCount"),
			jen.Id("errorCount").Qual("github.com/go-kit/kit/metrics", "Counter"),
			jen.Id("requestLatency").Qual("github.com/go-kit/kit/metrics", "Histogram")).
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
						g.Id("mw").Dot("requestCount").Dot("With").
							Call(jen.Lit("method"), jen.Lit(method.Name())).
							Dot("Add").Call(jen.Lit(1))

						g.Id("mw").Dot("requestLatency").Dot("With").
							Call(jen.Lit("method"), jen.Lit(method.Name())).
							Dot("Observe").Call(jen.Id("time.Since").Call(jen.Id("begin")).Dot("Seconds").Call())

						if errorResult := method.ErrorResult(); errorResult != nil {
							g.If(jen.Id(errorResult.Name()).Op("!=").Nil()).
								Block(jen.Id("mw").Dot("errorCount").Dot("With").
									Call(jen.Lit("method"), jen.Lit(method.Name())).
									Dot("Add").Call(jen.Lit(1)))
						}
					}).
					Call(jen.Qual("time", "Now").Call())
				g.Return(
					jen.Id("mw").Dot("next").Dot(method.Name()).CallFunc(func(g *jen.Group) {
						for _, param := range method.Params() {
							g.Id(param.Name())
						}
					}),
				)
			})
	}

	return nil
}
