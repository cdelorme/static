package staticmd

import (
	"html/template"
)

type Page struct {
	Name    string
	Version string
	Nav     []Navigation
	Depth   string
	Content template.HTML
}
