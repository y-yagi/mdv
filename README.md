# mdv

The `mdv` is a simple Markdown Viewer.

## Usage

Run `mdv <filename>` at terminal and access to http://localhost:8888/ .

The content automatically refreshes when the file is updated. You don't need to reload a page.

If you point local files in a markdown file, please specify a directory to `dir` option that local files exist.

```bash
mdv -dir . <filename>
```

If you want to customize a style of rendering, please specify a CSS file that describes the styles to CLI.

```bash
mdv -css <CSS filename> <filename>
```

## Installation

Download files from [GitHub release page](https://github.com/y-yagi/mdv/releases).
