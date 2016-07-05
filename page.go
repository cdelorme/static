package staticmd

import "html/template"

type page struct {
	Name    string
	Version string
	Nav     []navigation
	Depth   string
	Content template.HTML
}
