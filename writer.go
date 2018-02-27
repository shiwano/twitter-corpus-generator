package main

import (
	"io"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type tweetWriter struct {
	tweets  *lumberjack.Logger
	replies *lumberjack.Logger
}

func newTweetWriter(tweetFile, replyFile fileConfig) *tweetWriter {
	return &tweetWriter{
		tweets: &lumberjack.Logger{
			Filename: tweetFile.Path,
			MaxSize:  tweetFile.MaxSize,
		},
		replies: &lumberjack.Logger{
			Filename: replyFile.Path,
			MaxSize:  replyFile.MaxSize,
		},
	}
}

func (w *tweetWriter) writeTweet(t tweet) error {
	_, err := io.WriteString(w.tweets, t.text+"\n")
	return err
}

func (w *tweetWriter) writeReply(r reply) error {
	_, err := io.WriteString(w.replies, r.inReplyTo.text+"\t"+r.tweet.text+"\n")
	return err
}
