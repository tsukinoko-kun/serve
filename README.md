# serve

[![Test](https://github.com/tsukinoko-kun/serve/actions/workflows/test.yml/badge.svg)](https://github.com/tsukinoko-kun/serve/actions/workflows/test.yml)

Run `serve` to start a webserver hosting the content of a directory.

## Install

### Go install

```bash
go install github.com/tsukinoko-kun/serve@latest
```

### Homebrew

```sh
brew tap tsukinoko-kun/tap
brew install tsukinoko-kun/tap/serve
```

## Usage

```bash
serve [directory] [flags...]
```

## Flags

- `--md`

  Compile Markdown files to HTML

- `--port` `-p` int

  Port to listen on (default is a random port)

- `--verbose` `-v`

  Verbose output
