package service

import (
	"strings"

	parser "github.com/jackwakefield/go-parser"
)

type Method interface {
	Name() string
	LocalName() string
	GlobalName() string
	Params() []TypeValue
	Results() []TypeValue
	ErrorResult() TypeValue
}

type method struct {
	item    *parser.GoMethod
	params  []TypeValue
	results []TypeValue
}

func newMethod(item *parser.GoMethod) Method {
	return &method{
		item: item,
	}
}

func (m *method) Name() string {
	return m.item.Name
}

func (m *method) LocalName() string {
	name := m.Name()

	if len(name) > 0 {
		lowercase := strings.ToLower(string(name[0]))

		if len(name) == 1 {
			name = lowercase
		} else if len(name) >= 1 {
			name = lowercase + string(name[1:])
		}
	}

	return name
}

func (m *method) GlobalName() string {
	return strings.Title(m.Name())
}

func (m *method) Params() []TypeValue {
	if m.params == nil {
		m.params = make([]TypeValue, 0, len(m.item.Params))

		for _, param := range m.item.Params {
			m.params = append(m.params, newTypeValue(param))
		}
	}

	return m.params
}

func (m *method) Results() []TypeValue {
	if m.results == nil {
		m.results = make([]TypeValue, 0, len(m.item.Params))

		for _, param := range m.item.Results {
			m.results = append(m.results, newTypeValue(param))
		}
	}

	return m.results
}

func (m *method) ErrorResult() TypeValue {
	for _, item := range m.Results() {
		if item.IsErr() {
			return item
		}
	}

	return nil
}
