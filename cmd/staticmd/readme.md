
# staticmd

This is the cli implementation and interface to the [staticmd](https://github.com/cdelorme/staticmd) library.

It exposes the libraries functionality and handles inputs and outputs through command line.


## alternatives

There are other tools in multiple languages that already provide full-featured static generated assets with support for far more complex use cases:

- [node; harp.js](http://harpjs.com/).
- [golang; hugo](http://gohugo.io/)


## sales pitch

While the original point of this project was self-education, it still serves as a simple and compact shippable utility at only `834` lines of code (including tests and comments).  With recent updates it is now much cleaner and fully tested.

It has a minimal dependency footprint:

- [blackfriday](https://github.com/russross/blackfriday)
- [go-log](https://github.com/cdelorme/go-log)
- [go-maps](https://github.com/cdelorme/go-maps)
- [go-option](https://github.com/cdelorme/go-option)


## design

This project assumes some defaults behaviors:

- it tries all files even if it fails to read one or more of them, but it will produce a valid exit code
- it assumes the current working directory (eg. where it is run from) is the input path
- it assumes `public/` for the output path
- it assumes absolute paths by default
- it assumes a page per file by default

_Most of these default behaviors can be overridden with command line parameters._

The only required parameter is the path to a template file.  _An [example](template.tmpl) has been included with the project._

If run within a repository it will attempt to grab a short-hash for the version, otherwise it generates a unix timestamp.

It assumes the index files are `index.html`.  _If no match is found then no table of contents is created and no references to that folder will be created at the parent either even if it contains other markdown files._

Currently it attempts to generate multiple pages in parallel.


## usage

Install the utility:

    go get -u github.com/cdelorme/staticmd/...

The utility has builtin help to assist with using it:

	staticmd help

_It uses the [`go-log` package](https://github.com/cdelorme/go-log), which means you can enable logging by setting `LOG_LEVEL=debug` (or any other valid log level)._

You can use it to generate a single-page document that can be printed as a PDF and distributed with ease:

    staticmd -t template.tmpl -s -c src/

_This will generate a single file inside `src/`._


## testing

This code has unit tests that can easily be executed:

	LOG_LEVEL=silent go test -v -race -cover
