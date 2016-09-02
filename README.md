# Html2pdf

A CLI tool for generating PDF from HTML by using Lua script.

Html2pdf uses [wkhtmltopdf](http://wkhtmltopdf.org/) to generate PDF. Actually, Html2pdf is designed as a wrapper of the wkhtmltopdf command. But Html2pdf bundles the wkhtmltopdf binary, so you don't need to install wkhtmltopdf. Html2pdf works well in a stand-alone.

Table of Contents

* [Installation](#installation)
* [Getting Started](#getting-started)
* [Configuration](#configuration)
  * [Generate PDF](#generate-pdf)
  * [Generate PDF from URL](#generate-pdf-from-url)
  * [Multiple Pages](#multiple-pages)
  * [Change Output File](#change-output-file)
  * [Add Cover](#add-cover)
  * [Options](#options)
  * [Variables](#variables)
  * [Write Complex Config](#write-complex-config)
  * [DSL Syntax](dsl-syntax)
* [Developing Html2pdf](developing-html2pdf)
* [TODO](#todo)
* [Author](#author)
* [License](#license)
* [See Also](#see-also)

## Installation

Html2pdf is provided as a single binary. You can download it and drop it in your $PATH.

[Download latest version](https://github.com/kohkimakimoto/html2pdf/releases/latest)

## Getting Started

Create Lua script `hello.lua` as the following:

```lua
local html2pdf = require "html2pdf"

local hello = html2pdf.pdf "hello.pdf"
hello.options = {
    page_size = "A4",
}
hello.pages = {
    input_content = "hello world!"
}
```

Run it:

```
$ html2pdf hello.lua
==> Starting html2pdf...
==> Loaded 1 pdf config.
==> Evaluating hello.pdf
    output_file: hello.pdf
==> Complete!
```

You will get `hello.pdf` file.

## Configuration

Configuration in the Html2pdf is written in Lua.

### Generate PDF

```lua
local html2pdf = require "html2pdf"

local example = html2pdf.pdf "example.pdf"
example.pages = {
    input_content = "hello world!"
}
```

`html2pdf` is a lua module of the Html2pdf and `html2pdf.pdf` function defines a config of generating PDF file.

### Generate PDF from URL

By using `input` key, you can use a URL instead of html content.

```lua
local html2pdf = require "html2pdf"

local example = html2pdf.pdf "example.pdf"
example.pages = {
    input = "https://github.com/kohkimakimoto/html2pdf"
}
```

### Multiple Pages

You can set mulitple pages.

```lua
example.pages = {
    { input = "https://github.com/kohkimakimoto/html2pdf" },
    { input = "https://github.com/kohkimakimoto/cof" },
}
```

### Change Output File

You can change output file name.

```lua
example.output_file = "output.pdf"
```

### Add Cover

```lua
example.cover = {
    input_content = "<b>title</b>"
    -- or
    -- input = "https://..."
}
```

### Options

You can set wkhtmltopdf options.

```lua
example.options = {
    page_size = "A4",
    margin_left = 5,
    margin_top = 5,
    margin_bottom = 5,
    margin_right = 5,
    orientation = "Landscape",
    -- etc...
}
```

See also: [wkhtmltopdf docs](http://wkhtmltopdf.org/docs.html)

### Variables

You can input variables to a config by `-var` and `-var-file` option.
The variables can be read in a config by `var` global variable.

Example: create `example.lua`.

```lua
local html2pdf = require "html2pdf"

local example = html2pdf.pdf "example.pdf"
example.output_file = var.output_file
example.pages = {
    input = "https://github.com/kohkimakimoto/html2pdf"
}
```

Run html2pdf with `-var`.

```
$ html2pdf example.lua -var='{"output_file": "foo.pdf"}'
```

### Write Complex Config

For writing more complex configuration, Html2pdf bundles the following lua modules.

* `json`: [layeh/gopher-json](https://github.com/layeh/gopher-json).
* `fs`: [kohkimakimoto/gluafs](https://github.com/kohkimakimoto/gluafs).
* `yaml`: [kohkimakimoto/gluayaml](https://github.com/kohkimakimoto/gluayaml).
* `template`: [kohkimakimoto/gluatemplate](https://github.com/kohkimakimoto/gluatemplate).
* `makrdown`: [kohkimakimoto/makrdown](https://github.com/kohkimakimoto/makrdown).
* `env`: [kohkimakimoto/gluaenv](https://github.com/kohkimakimoto/gluaenv).
* `http`: [cjoudrey/gluahttp](https://github.com/cjoudrey/gluahttp).
* `re`: [yuin/gluare](https://github.com/yuin/gluare)
* `sh`:[otm/gluash](https://github.com/otm/gluash)

For instance, you can generate PDF from markdown text. See the [example](_example).

### DSL Syntax

Html2pdf supports to write code as DSL style. See the following example.

```lua
pdf "hello.pdf" {
    pages = {
        input_content = "hello world!"
    },
}
```

The `pdf` function is an alias of `html2pdf.pdf` method. So this is equivalent the following code:

```lua
local html2pdf = require "html2pdf"

local hello = html2pdf.pdf "hello.pdf"
hello.pages = {
    input_content = "hello world!"
}
```

## Developing Html2pdf

Requirements

* Go 1.7 or later (my development env)
* [Gom](https://github.com/mattn/gom)

Installing dependences

```
$ make deps
```

Building dev binary.

```
$ make build_bindata
$ make
```

Building distributed binaries.


```
$ make build_bindata
$ make dist
```

## TODO

* support windows
* support page options
* support TOC

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>

## License

The MIT License (MIT)

## See Also

* [wkhtmltopdf](http://wkhtmltopdf.org/)
* [SebastiaanKlippert/go-wkhtmltopdf](https://github.com/SebastiaanKlippert/go-wkhtmltopdf)
