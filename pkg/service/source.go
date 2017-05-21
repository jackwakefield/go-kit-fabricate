package service

import "github.com/jackwakefield/go-parser"

type Source interface {
	Package() string
	Service() Service
}

type source struct {
	file    *parser.GoFile
	service Service
}

func ParseFile(path string) (Source, error) {
	file, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}
	return &source{file: file}, nil
}

func (s *source) Package() string {
	return s.file.Package
}

func (s *source) Service() Service {
	if s.service == nil {
		for _, item := range s.file.Interfaces {
			if item.Name == "Service" {
				s.service = newService(item)
				break
			}
		}
	}
	return s.service
}
