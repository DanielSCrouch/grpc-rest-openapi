package testdata

import "demoserver/api"

var TestCell = api.Cell{
	Identity: &api.Identifier{Uuid: "foo-cell"},
	Status:   "foo-status",
}
