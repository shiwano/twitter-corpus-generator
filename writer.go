package main

import (
	"io"
	"strings"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type tweetWriter struct {
	tweets  *lumberjack.Logger
	dialogs *lumberjack.Logger
}

func newTweetWriter(tweetFile, dialogFile fileConfig) *tweetWriter {
	return &tweetWriter{
		tweets: &lumberjack.Logger{
			Filename: tweetFile.Path,
			MaxSize:  tweetFile.MaxSize,
		},
		dialogs: &lumberjack.Logger{
			Filename: dialogFile.Path,
			MaxSize:  dialogFile.MaxSize,
		},
	}
}

func (w *tweetWriter) writeTweet(t tweet) error {
	_, err := io.WriteString(w.tweets, t.text+"\n")
	return err
}

func (w *tweetWriter) writeDialog(d dialog) error {
	texts := make([]string, len(d.tweets))
	for i, t := range d.tweets {
		texts[len(d.tweets)-(i+1)] = t.text
	}
	s := strings.Join(texts, "\t") + "\n"
	_, err := io.WriteString(w.dialogs, s)
	return err
}
