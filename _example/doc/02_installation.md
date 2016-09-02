# Installation

Html2pdf is provided as a single binary. You can download it and drop it in your $PATH.

[Download latest version](https://github.com/kohkimakimoto/html2pdf/releases/latest)

# Getting Started

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
