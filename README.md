# mdv

The `mdv` is a simple Markdown Viewer.

## Usage

Run `mdv <filename>` at terminal and access to http://localhost:8888/ .

The content automatically refreshes when the file is updated. You don't need to reload a page.

If you point local files in a markdown file, please specify a directory to `dir` option that local files exist.

```bash
mdv -dir . <filename>
```

### Style

`mdv` uses [github-markdown-css](https://github.com/sindresorhus/github-markdown-css) by default.  If you want to use an original style, please specify a `css` option. `mdv` only use the CSS file if `css` option is specified.

```bash
mdv -css <CSS filename> <filename>
```

## Installation

Download files from [GitHub release page](https://github.com/y-yagi/mdv/releases).
