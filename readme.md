
# staticmd

This is a go library written to provide static generation of html from markdown with a template.

The intended function is to make it easy to convert large folders of documentation in bulk to explorable web content or a single-page document for printing as a PDF and redistributing.

_This project was used as a means of self-education for learning the go programming language and tooling._


## design

The design focuses on three key outcomes:

- absolute paths for a website
- relative paths for local content
- single-file with a table of contents

It depends on:

- markdown
- a template

_An [example template](cmd/staticmd/template.tmpl) has been provided._

It automates two convenience outputs:

- automatically generated index navigation
- asset versions for cache busting


### template

A user-defined template affords you flexibility and control.

A limited subset of variables are provided for your use:

- Depth
- Version
- Nav
- Content

The `.Depth` property is used when generating paths to global resources such as javascript, style sheets, and images.

The `.Version` property allows you to address cached resources such as javascript, style sheets, and resources javascript may ask for.  _It is highly efficient as the cache will immediately rebuild using the new version._

The `.Nav` is the table-of-contents produced in single-page or "book" mode, and is empty otherwise.

The `.Content` is the file contents converted to html, and will have a table of contents automatically prepended to any `index.html` file deeper than the root of the output folder.

_Any assets hosted on a CDN can be given full paths, and it is recommended to host binary resources (such as images) and any javascript and style sheets there as well._


## usage

To import the library:

	import "github.com/cdelorme/staticmd"

To install the cli utility:

	go get -u github.com/cdelorme/staticmd/...

_The ellipses will install all packages below the main path._
