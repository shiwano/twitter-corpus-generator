# twitter-corpus-generator

> :bird: twitter-corpus-generator saves sample tweets and dialog tweets from Twitter, and rotates automatically output files based on [lumberjack](https://github.com/natefinch/lumberjack)'s logic.

## Getting Started

Execute the following commands, so you will get corpus data in `data` directory.

```sh
$ mkdir twitter-corpus && cd twitter-corpus
$ curl -L -o config.toml https://raw.githubusercontent.com/shiwano/twitter-corpus-generator/master/config.example.toml
$ vi config.toml # Fill your twitter application tokens.
$ go get -u github.com/shiwano/twitter-corpus-generator
$ twitter-corpus-generator # If `-verbose` option is given, it shows tweets on stdout.
```

If you does not install Go, get the binary from [Cloud Gox](https://gox.jpillora.com/).
