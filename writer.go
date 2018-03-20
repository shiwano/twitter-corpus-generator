package main

import (
	"io"
	"strings"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type writer struct {
	tweets  *lumberjack.Logger
	dialogs *lumberjack.Logger
}

func newWriter(tweetFile, dialogFile fileConfig) *writer {
	return &writer{
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

func (w *writer) writeTweet(t tweet) error {
	_, err := io.WriteString(w.tweets, t.text+"\n")
	return err
}

func (w *writer) writeDialog(d dialog) error {
	texts := make([]string, len(d.tweets))
	for i, t := range d.tweets {
		texts[len(d.tweets)-(i+1)] = t.text
	}
	s := strings.Join(texts, "\t") + "\n"
	_, err := io.WriteString(w.dialogs, s)
	return err
}
