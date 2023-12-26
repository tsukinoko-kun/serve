# serve

[![Test](https://github.com/Frank-Mayer/serve/actions/workflows/test.yml/badge.svg)](https://github.com/Frank-Mayer/serve/actions/workflows/test.yml)

Run `serve` to start a webserver hosting the content of a directory.

## Install

```bash
go install github.com/Frank-Mayer/serve/cmd/serve@latest
```

## Flags

- `dir` string

  Path to the directory to serve (default ".")

- `md`

   Compile Markdown files to HTML

- `port` int

  Port to listen on (default 8080)

- `verbose`

  Verbose output

- `version`

  Print version
