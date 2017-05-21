package generator

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/tools/imports"

	"github.com/dave/jennifer/jen"
	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
)

type fileGenerator struct {
	generator func(file *jen.File) error
	source    service.Source
}

func newFileGenerator(source service.Source) *fileGenerator {
	return &fileGenerator{
		source: source,
		generator: func(file *jen.File) error {
			return nil
		},
	}
}

func (g *fileGenerator) GenerateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return g.Generate(path, file)
}

func (g *fileGenerator) Generate(path string, writer io.Writer) error {
	file := jen.NewFile(g.source.Package())
	if err := g.generator(file); err != nil {
		return err
	}

	buffer := &bytes.Buffer{}
	if err := file.Render(buffer); err != nil {
		return err
	}

	src, err := ioutil.ReadAll(buffer)
	if err != nil {
		return err
	}

	res, err := imports.Process(path, src, nil)
	if err != nil {
		return err
	}

	_, err = writer.Write(res)
	return err
}
