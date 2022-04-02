package testdata

import "demoserver/api"

var TestCell = api.Cell{
	Identity: &api.Identity{Identity: "foo-cell"},
	Status:   "foo-status",
}
