package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

const version = "1.0.0"

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

	writer := newTweetWriter(c.TweetFile, c.ReplyFile)

	stream, err := newTweetStream(context.Background(), c.Language, c.ConsumeKey, c.ConsumeKeySecret, c.AccessToken, c.AccessTokenSecret)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for reply := range stream.replies {
			if verboseFlag {
				log.Println("Reply1: ", reply.inReplyTo.text)
				log.Println("Reply2: ", reply.tweet.text)
			}
			writer.writeReply(reply)
		}
	}()

	for tweet := range stream.tweets {
		if verboseFlag {
			log.Println("Tweet: ", tweet.text)
		}
		writer.writeTweet(tweet)
	}
}
