
[![GoDoc](https://godoc.org/github.com/cdelorme/static?status.svg)](https://godoc.org/github.com/cdelorme/static)

# [static](https://github.com/cdelorme/static)

An abstraction ontop of [blackfriday](https://godoc.org/github.com/russross/blackfriday) for the purpose of accepting a folder of files and generating html in lexical order with built-in templates.

_The primary objective of this project was originally self-education._  The result is a sub-par utility with minimal features.  The only benefits it offers are a significantly smaller code base, and no meta-data dependencies.


## usage

To import the library:

	import "github.com/cdelorme/static"


## notes

Relative path support has been removed because raw markdown is intended to be readable.

Automatic navigation has been removed from the web solution, since the requirements vary by website and are entirely different when generating a book.  _Use the template override feature to create your own._

The code makes no assumptions about what index name is used, since that is entirely controlled by the web server.

The library is not concurrently safe, because there are zero benefits to running it concurrently.  Everything is bottlenecked at the hard drive, and that cannot be addressed without proper buffered solutions to both markdown and template parsing.

It uses [go-bindata](https://github.com/jteeuwen/go-bindata) to embed default templates, which have been committed to the project since `go generate` is not possible to do from `go get`.

No efforts have been made to optimize re-execution around existing files, _but it would be possible to compare the markdown file modified time against the modified time of existing html files to reduce overhead in the future._

If two files with alternative markdown extensions but identical base names exist, the first match is the only one that will be parsed into an html file.

The book mode provides automatically generated navigation using javascript.

Any absolute links in book mode will not function as desired.

All tests have been (re) written using a black-box approach, where only publicly exposed functions and properties are modified.


# references

- [blackfriday](https://godoc.org/github.com/russross/blackfriday)
- [function call from template](http://stackoverflow.com/questions/10200178/call-a-method-from-a-go-template)
- [buffer blackfriday output](http://grokbase.com/t/gg/golang-nuts/142spmv4fe/go-nuts-differences-between-os-io-ioutils-bufio-bytes-with-buffer-type-packages-for-file-reading)
- [github markdown file extensions](https://github.com/github/markup/blob/b865add2e053f8cea3d7f4d9dcba001bdfd78994/lib/github/markups.rb#L1)
- [javascript queryselector](http://caniuse.com/#feat=queryselector)
- [chaining writers in go](https://medium.com/@skdomino/writing-on-the-train-chaining-io-writers-in-go-1b39e07f71c9)
