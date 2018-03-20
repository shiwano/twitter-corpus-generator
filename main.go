package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

const version = "1.0.1"

func main() {
	var (
		verboseFlag bool
		versionFlag bool
	)
	flag.BoolVar(&verboseFlag, "verbose", false, "show tweets on stdout")
	flag.BoolVar(&versionFlag, "v", false, "show version")
	flag.BoolVar(&versionFlag, "version", false, "show version")
	flag.Parse()

	if versionFlag {
		fmt.Println(version)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime)

	c, err := loadConfig("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	writer := newTweetWriter(c.TweetFile, c.DialogFile)

	stream, err := newTweetStream(context.Background(), c.Language, c.ConsumeKey, c.ConsumeKeySecret, c.AccessToken, c.AccessTokenSecret)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for d := range stream.dialogs {
			if len(d.tweets) >= 3 {
				for i := 0; i < len(d.tweets); i++ {
					log.Printf("Dialog%d: %s", i, d.tweets[len(d.tweets)-(i+1)].text)
				}
			}
			writer.writeDialog(d)
		}
	}()

	for tweet := range stream.tweets {
		if verboseFlag {
			log.Println("Tweet: ", tweet.text)
		}
		writer.writeTweet(tweet)
	}
}
