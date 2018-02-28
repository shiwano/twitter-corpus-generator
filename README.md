# twitter-corpus-generator

Twitter corpus generator written in Go, it generates sample streaming tweets and in-reply-to tweets data with automatic file rotation.

## Getting Started

Execute the following commands, so you will get corpus data in `data` directory.

```sh
$ mkdir twitter-corpus && cd twitter-corpus && curl -L -o config.toml https://raw.githubusercontent.com/shiwano/twitter-corpus-generator/master/config.example.toml
$ vi config.toml # Fill your twitter application tokens.
$ go get -u github.com/shiwano/twitter-corpus-generator
$ twitter-corpus-generator # If `-verbose` option is given, it shows tweets on stdout.
```
