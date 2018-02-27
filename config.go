package main

import "github.com/BurntSushi/toml"

type fileConfig struct {
	Path    string `toml:"path"`
	MaxSize int    `toml:"max_size"` // megabytes
}

type config struct {
	ConsumeKey        string `toml:"consume_key"`
	ConsumeKeySecret  string `toml:"consume_key_secret"`
	AccessToken       string `toml:"access_token"`
	AccessTokenSecret string `toml:"access_token_secret"`

	Language string `toml:"language"`

	TweetFile fileConfig `toml:"tweet_file"`
	ReplyFile fileConfig `toml:"reply_file"`
}

func loadConfig(path string) (*config, error) {
	var c *config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, err
	}
	return c, nil
}
