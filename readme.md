
# staticmd

This is a golang project for a command that generates static html output from markdown documents and a template.  It's purpose is to make it easy to convert large folders in bulk, as well as single-page support.  It works for simple static content websites, as well as documentation.


## alternatives

Other tools exist, not limited to the golang language:

- [node; harp.js](http://harpjs.com/).
- [golang; hugo](http://gohugo.io/)

_My project was for educational purposes, and to provide a much more basic utility than the above options._


## sales pitch

The aim of this tool is to provide a simple command with sensible flags that will generate:

- a full website
- a relative-pathed local readme
- a single file with a table of contents

It depends on:

- a user-supplied template
- a bunch of markdown files

It supplies the template with convenient output, such as:

- top level navigation
- asset versioning

Its options include:

- template path
- input path
- output path
- single-page output
- relative links
- debug mode
- cpu profiling


**Behavior:**

The template path is the only required parameter, and an example will be included with the project source.

It assumes the input path is the execution directory.

It assumes an output path of `public/` relative to the execution path.

Single page output will combine all files in the input path into a single output; something akin to a "book".  When this flag is selected the full table of contents is listed first; depth is applied based on the folder layout, and in the order which it was parsed.

By default the application will assume absolute paths starting at the `output` directory.  If you use the `relative` flag it will use relative-path file names and rely on a depth parameter in the template.

If run within a repository it will attempt to grab the short-hash from git as a version, otherwise it will use a unix timestamp.

It assumes the index will match `index.html`.  If no match is found no table of contents are created, and no references to that folders contents will be created, even if it contains markdown files.

For multiple page generation it utilizies concurrency.  _When run in single-page mode it cannot build concurrently._


**It does not come with:**

- abstractions
- interfaces
- unit tests


## usage

Install the command:

    go get github.com/cdelorme/staticmd

In this example we can generate a single-page document for easier printing and sharing:

    staticmd -t template.tmpl -s -c src/

**The above command will generate a single file from the template using files inside `src/`.**

_For further details on command line operation, use the `--help` flag._


### template file

A single template file can be used to generate all pages; even single-page output.

The following variables are provided to the template file:

- Depth
- Version
- Nav
- Content

The `.Depth` property is for links you supply, such as to css and js assets.

The `.Version` tag allows you to prevent caching of changed assets, such as css and js, but can also be supplied.

The `.Nav` is a set of top-level links only.  In single-page mode the navigation is omitted.

The `.Content` is the file contents converted to html.  Any `index.html` files deeper than the output folder will automatically have a table of contents prepended to them.

_Any binary assets such as images should be hosted on a CDN and have full paths.  If you need to supply a full offline version then use the single-page mode with this tool and print to pdf._
