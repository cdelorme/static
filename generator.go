package staticmd

import (
	"bufio"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

var readfile = ioutil.ReadFile
var create = os.Create
var mkdirall = os.MkdirAll
var parseFiles = template.ParseFiles
var walk = filepath.Walk

type executor interface {
	Execute(io.Writer, interface{}) error
}

type Generator struct {
	Book         bool   `json:"book,omitempty"`
	Input        string `json:"input,omitempty"`
	Output       string `json:"output,omitempty"`
	Relative     bool   `json:"relative,omitempty"`
	TemplateFile string `json:"template,omitempty"`

	L logger

	version  string
	pages    []string
	template executor
}

func (g *Generator) ior(path string) string {
	return strings.TrimSuffix(strings.Replace(path, g.Input, g.Output, 1), filepath.Ext(path)) + ".html"
}

func (g *Generator) depth(path string) string {
	if g.Relative {
		if rel, err := filepath.Rel(filepath.Dir(path), g.Output); err == nil {
			return rel + string(os.PathSeparator)
		}
	}
	return ""
}

func (g *Generator) walk(path string, file os.FileInfo, err error) error {
	if file != nil && file.Mode().IsRegular() && file.Size() > 0 && isMarkdown(path) {
		g.pages = append(g.pages, path)
	}
	return err
}

func (g *Generator) multi() error {
	navi := make(map[string][]navigation)
	var err error

	for i, _ := range g.pages {
		out := g.ior(g.pages[i])
		dir := filepath.Dir(g.ior(out))
		nav := navigation{}

		if filepath.Dir(out) != g.Output && strings.ToLower(basename(out)) == "index" {
			nav.Title = basename(dir)
			if g.Relative {
				nav.Link = filepath.Join(strings.TrimPrefix(dir, filepath.Dir(dir)+string(os.PathSeparator)), filepath.Base(out))
			} else {
				nav.Link = strings.TrimPrefix(dir, g.Output) + string(os.PathSeparator)
			}
			dir = filepath.Dir(dir)
		} else {
			nav.Title = basename(out)
			if g.Relative {
				nav.Link = strings.TrimPrefix(out, filepath.Dir(out)+string(os.PathSeparator))
			} else {
				nav.Link = strings.TrimPrefix(out, g.Output)
			}
		}

		if _, ok := navi[dir]; !ok {
			navi[dir] = make([]navigation, 0)
			if ok, _ := exists(dir); !ok {
				if err = mkdirall(dir, 0770); err != nil {
					g.L.Error("failed to create path: %s, %s", dir, err)
				}
			}
		}

		navi[dir] = append(navi[dir], nav)
	}

	for _, p := range g.pages {
		var markdown []byte
		if markdown, err = readfile(p); err != nil {
			g.L.Error("failed to read file: %s, %s", p, err)
			return err
		}

		out := g.ior(p)
		dir := filepath.Dir(out)
		page := page{
			Name:    basename(p),
			Version: g.version,
			Nav:     navi[g.Output],
			Depth:   g.depth(out),
		}

		if dir != g.Output && strings.ToLower(basename(p)) == "index" {
			toc := "\n## Table of Contents:\n\n"
			for i, _ := range navi[dir] {
				toc = toc + "- [" + navi[dir][i].Title + "](" + navi[dir][i].Link + ")\n"
			}
			g.L.Debug("table of contents for %s, %s", out, toc)
			markdown = append([]byte(toc), markdown...)
		}

		page.Content = template.HTML(blackfriday.MarkdownCommon(markdown))

		var f *os.File
		if f, err = create(out); err != nil {
			g.L.Error("%s\n", err)
			return err
		}
		defer f.Close()

		fb := bufio.NewWriter(f)
		defer fb.Flush()

		if err = g.template.Execute(fb, page); err != nil {
			g.L.Error("%s\n", err)
		}
	}

	return err
}

func (g *Generator) single() error {
	content := make([]byte, 0)
	toc := "\n"
	previous_depth := 0
	var err error

	for _, p := range g.pages {
		shorthand := strings.TrimPrefix(p, g.Input+string(os.PathSeparator))
		depth := strings.Count(shorthand, string(os.PathSeparator))
		if depth > previous_depth {
			toc = toc + strings.Repeat("\t", depth-1) + "- " + basename(filepath.Dir(p)) + "\n"
		}
		anchor := strings.Replace(shorthand, string(os.PathSeparator), "-", -1)
		toc = toc + strings.Repeat("\t", depth) + "- [" + basename(p) + "](#" + anchor + ")\n"

		var markdown []byte
		if markdown, err = readfile(p); err != nil {
			g.L.Error("failed to read file: %s (%s)", p, err)
			continue
		}

		markdown = append([]byte("\n<a id='"+anchor+"'/>\n\n"), markdown...)
		markdown = append(markdown, []byte("\n[back to top](#top)\n\n")...)
		content = append(content, markdown...)
		previous_depth = depth
	}

	content = append([]byte(toc), content...)

	if ok, _ := exists(g.Output); !ok {
		if err = mkdirall(g.Output, 0770); err != nil {
			g.L.Error("failed to create path: %s (%s)", g.Output, err)
			return err
		}
	}

	page := page{
		Version: g.version,
		Content: template.HTML(blackfriday.MarkdownCommon(content)),
	}
	out := filepath.Join(g.Output, "index.html")

	var f *os.File
	if f, err = create(out); err != nil {
		g.L.Error("%s\n", err)
		return err
	}
	defer f.Close()

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	if err = g.template.Execute(fb, page); err != nil {
		g.L.Error("%s\n", err)
	}

	return err
}

func (g *Generator) Generate() error {
	var err error
	if g.template, err = parseFiles(g.TemplateFile); err != nil {
		g.L.Error("%s\n", err)
		return err
	}

	g.version = version(g.Input)
	g.Input, _ = filepath.Abs(g.Input)
	g.Output, _ = filepath.Abs(g.Output)
	g.Input = filepath.Clean(g.Input)
	g.Output = filepath.Clean(g.Output)

	if err := walk(g.Input, g.walk); err != nil {
		g.L.Error("%s\n", err)
		return err
	}
	g.L.Debug("generator state: %+v", g)

	if g.Book {
		return g.single()
	}
	return g.multi()
}
