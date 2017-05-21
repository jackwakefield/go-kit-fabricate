package main

import (
	"os"
	"path/filepath"

	"github.com/jackwakefield/go-kit-fabricate/pkg/generator"
	"github.com/jackwakefield/go-kit-fabricate/pkg/service"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "fabricate"
	app.Usage = "generate the logging and instrumentation middleware, transport, endpoints and client for go-kit"
	app.UsageText = "fabricate [global options] [file path]"

	app.Before = func(c *cli.Context) error {
		if !c.Args().Present() {
			return cli.ShowAppHelp(c)
		}
		return nil
	}

	app.Action = func(c *cli.Context) error {
		path := c.Args().First()
		s, err := service.ParseFile(path)
		if err != nil {
			return err
		}

		dir := filepath.Dir(path)
		g := generator.NewGenerator(s, dir)

		if err := g.GenerateMiddleware(); err != nil {
			return err
		}

		if err := g.GenerateLogging(); err != nil {
			return err
		}

		if err := g.GenerateInstrumenting(); err != nil {
			return err
		}

		if err := g.GenerateEndpoints(); err != nil {
			return err
		}

		return nil
	}

	app.Run(os.Args)
}
