package service

import parser "github.com/jackwakefield/go-parser"

type Service interface {
	Methods() []Method
}

type service struct {
	item    *parser.GoInterface
	methods []Method
}

func newService(item *parser.GoInterface) Service {
	s := &service{item: item}

	s.methods = make([]Method, 0, len(item.Methods))
	for _, method := range item.Methods {
		s.methods = append(s.methods, newMethod(method))
	}

	return s
}

func (s *service) Methods() []Method {
	return s.methods
}
