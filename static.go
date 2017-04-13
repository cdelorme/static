// This package provides a utility to parse markdown files in lexical order
// as a single document or individual pages, with no additional requirements.
//
// It includes default templates that are embedded, but can be directed to a
// separate file granting better control for more complex use-cases.
//
// Template parameters are simple, and include Title, Content, and Version;
// both the Version and Title can be changed.  If in web mode, an additional
// property called Name will be set to the basename of the file.
package static

// List of extensions matching the github parser, but with an inversed order
// so that we pick the shortest extension first.
var extensions = []string{".md", ".mkd", ".mkdn", ".mdown", ".markdown"}
