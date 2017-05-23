package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
)

type endpointsGenerator struct {
	*fileGenerator
}

func newEndpointsGenerator(source service.Source) *endpointsGenerator {
	g := &endpointsGenerator{newFileGenerator(source)}
	g.generator = g.generate
	return g
}

func (g *endpointsGenerator) generate(file *jen.File) error {
	if err := g.generateEndpointStruct(file); err != nil {
		return err
	}

	if err := g.generateEndpoints(file); err != nil {
		return err
	}

	if err := g.generateMiddleware(file); err != nil {
		return err
	}

	return nil
}

func (g *endpointsGenerator) generateEndpointStruct(file *jen.File) error {
	service := g.source.Service()

	file.Type().Id("Endpoints").StructFunc(func(g *jen.Group) {
		for _, item := range service.Methods() {
			g.Id(fmt.Sprintf("%sEndpoint", item.Name())).Qual("github.com/go-kit/kit/endpoint", "Endpoint")
		}
	}).Line()

	return nil
}

func (g *endpointsGenerator) generateEndpoints(file *jen.File) error {
	service := g.source.Service()

	for _, method := range service.Methods() {
		requestName := fmt.Sprintf("%sRequest", method.LocalName())
		file.Type().Id(requestName).StructFunc(func(g *jen.Group) {
			for _, param := range method.Params() {
				if !param.IsContext() {
					g.Id(param.GlobalName()).Id(param.Type())
				}
			}
		}).Line()

		responseName := fmt.Sprintf("%sResponse", method.LocalName())
		file.Type().Id(responseName).StructFunc(func(g *jen.Group) {
			for _, result := range method.Results() {
				if !result.IsContext() {
					g.Id(result.GlobalName()).Id(result.Type())
				}
			}
		}).Line()

		file.Func().
			Params(jen.Id("e").Id("Endpoints")).
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
				g.Id("request").Op(":=").Id(requestName).Values(jen.DictFunc(func(d jen.Dict) {
					for _, param := range method.Params() {
						if !param.IsContext() {
							d[jen.Id(param.GlobalName())] = jen.Id(param.Name())
						}
					}
				}))

				g.List(jen.Id("response"), jen.Err()).Op(":=").Id(fmt.Sprintf("e.%sEndpoint", method.Name())).
					CallFunc(func(g *jen.Group) {
						hasContext := false
						for _, param := range method.Params() {
							if param.IsContext() {
								g.Id(param.Name())
								hasContext = true
								break
							}
						}
						if !hasContext {
							g.Qual("context", "Background").Call()
						}

						g.Id("request")
					})

				g.If(jen.Err().Op("!=").Nil()).Block(
					jen.ReturnFunc(func(g *jen.Group) {
						for _, result := range method.Results() {
							if result.IsErr() {
								g.Err()
							} else {
								g.Id(result.ZeroValue())
							}
						}
					}),
				)

				g.ReturnFunc(func(g *jen.Group) {
					for _, result := range method.Results() {
						g.Id("response").Assert(jen.Id(responseName)).Dot(result.GlobalName())
					}
				})
			}).
			Line()

		file.Func().
			Id(fmt.Sprintf("Make%sEndpoint", method.Name())).
			Params(jen.Id("s").Id("Service")).
			Params(jen.Qual("github.com/go-kit/kit/endpoint", "Endpoint")).
			Block(jen.Return().Func().
				Params(
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("request").Interface(),
				).
				Params(
					jen.Id("response").Interface(),
					jen.Err().Error(),
				).
				BlockFunc(func(g *jen.Group) {
					requestVar := fmt.Sprintf("%sReq", method.LocalName())

					g.Id(requestVar).
						Op(":=").
						Id("request").Assert(jen.Id(requestName))

					g.ListFunc(
						func(g *jen.Group) {
							for _, result := range method.Results() {
								g.Id(result.Name())
							}
						}).
						Op(":=").
						Id("s").Dot(method.Name()).
						CallFunc(func(g *jen.Group) {
							g.Id("ctx")

							for _, param := range method.Params() {
								if !param.IsContext() {
									g.Id(requestVar).Dot(param.GlobalName())
								}
							}
						})

					g.Return().List(
						jen.Id(responseName).Values(jen.DictFunc(func(d jen.Dict) {
							for _, result := range method.Results() {
								if !result.IsContext() {
									d[jen.Id(result.GlobalName())] = jen.Id(result.Name())
								}
							}
						})),
						jen.Nil(),
					)
				}),
			).
			Line()
	}

	return nil
}

func (g *endpointsGenerator) generateMiddleware(file *jen.File) error {
	file.Func().
		Id("EndpointInstrumentingMiddleware").
		Params(jen.Id("duration").Qual("github.com/go-kit/kit/metrics", "Histogram")).
		Params(jen.Qual("github.com/go-kit/kit/endpoint", "Middleware")).
		Block(jen.Return().Func().
			Params(jen.Id("next").Qual("github.com/go-kit/kit/endpoint", "Endpoint")).
			Params(jen.Qual("github.com/go-kit/kit/endpoint", "Endpoint")).
			Block(jen.Return().Func().
				Params(
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("request").Interface(),
				).
				Params(
					jen.Id("response").Interface(),
					jen.Err().Error(),
				).
				Block(
					jen.Defer().
						Func().
						Params(jen.Id("begin").Qual("time", "Time")).
						Block(
							jen.Id("duration").
								Dot("With").Call(jen.Lit("success"), jen.Qual("fmt", "Sprint").Call(jen.Id("err").Op("==").Nil())).
								Dot("Observe").Call(jen.Qual("time", "Since").Call(jen.Id("begin")).Dot("Seconds").Call()),
						).
						Call(jen.Qual("time", "Now").Call()).
						Line(),

					jen.Return(jen.Id("next").Call(jen.Id("ctx"), jen.Id("request"))),
				),
			),
		).
		Line()

	file.Func().
		Id("EndpointLoggingMiddleware").
		Params(jen.Id("logger").Qual("github.com/go-kit/kit/log", "Logger")).
		Params(jen.Qual("github.com/go-kit/kit/endpoint", "Middleware")).
		Block(jen.Return().Func().
			Params(jen.Id("next").Qual("github.com/go-kit/kit/endpoint", "Endpoint")).
			Params(jen.Qual("github.com/go-kit/kit/endpoint", "Endpoint")).
			Block(jen.Return().Func().
				Params(
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("request").Interface(),
				).
				Params(
					jen.Id("response").Interface(),
					jen.Err().Error(),
				).
				Block(
					jen.Defer().
						Func().
						Params(jen.Id("begin").Qual("time", "Time")).
						Block(
							jen.Id("logger").
								Dot("Log").
								Call(
									jen.Lit("error"),
									jen.Id("err"),
									jen.Lit("took"),
									jen.Qual("time", "Since").Call(jen.Id("begin")),
								),
						).
						Call(jen.Qual("time", "Now").Call()).
						Line(),
					jen.Return(jen.Id("next").Call(jen.Id("ctx"), jen.Id("request"))),
				),
			),
		).
		Line()

	return nil
}
