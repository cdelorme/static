
# smd

This is the cli implementation and interface to the [static](https://github.com/cdelorme/static) library for generating html from markdown.


## alternatives

There are far better tools written in a variety of languages which offer more features and better support.  _However they may also impose a runtime, special configuration, or special syntax_:

- [node; harp.js](http://harpjs.com/).
- [golang; hugo](http://gohugo.io/)

So, while the purpose of this project was self-education, what it provides is a very compact tool with minimal third party dependencies and no special requirements for configuration or syntax, wrapped in under 700 lines of code (including tests and comments).


## design

This project assumes the following behaviors:

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

_Logging is completely controlled by the `LOG_LEVEL` environment variable, which accepts standard syslog severities with a default of `error` (use `silent` to turn off output)._

You can use it to generate a single-page document that can be printed as a PDF and distributed with ease:

    staticmd -t template.tmpl -s -c src/

_This will generate a single file inside `src/`._


## testing

This code has unit tests that can easily be executed:

	LOG_LEVEL=silent go test -v -race -cover
