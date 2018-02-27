package main

import (
	"context"
	"flag"
	"fmt"
	"log"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "show tweets on stdout")
	flag.Parse()

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
			if verbose {
				fmt.Println("Reply1: ", reply.inReplyTo.text)
				fmt.Println("Reply2: ", reply.tweet.text)
			}
			writer.writeReply(reply)
		}
	}()

	for tweet := range stream.tweets {
		if verbose {
			fmt.Println("Tweet: ", tweet.text)
		}
		writer.writeTweet(tweet)
	}
}
