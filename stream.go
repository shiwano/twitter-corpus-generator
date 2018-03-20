package main

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const statusLookupMaxTweetCount = 100
const statusLookupMaxCallCount = 60
const statusLookupResetDuration = 16 * time.Minute

type stream struct {
	stream       *twitter.Stream
	client       *twitter.Client
	dialogTweets chan tweet

	dialogs chan dialog
	tweets  chan tweet
}

func (s *stream) makeDialogs(tweets []tweet) []dialog {
	tweetsByInReplyToStatusID := make(map[int64]tweet)
	inReplyToStatusIDs := make([]int64, len(tweets))
	for i, t := range tweets {
		inReplyToStatusIDs[i] = t.inReplyToStatusID
		tweetsByInReplyToStatusID[t.inReplyToStatusID] = t
	}
	lookupParams := &twitter.StatusLookupParams{TrimUser: twitter.Bool(true)}
	inReplyToRawTweets, _, err := s.client.Statuses.Lookup(inReplyToStatusIDs, lookupParams)
	if err != nil {
		log.Println(err)
		return nil
	}
	var dialogs []dialog
	for _, inReplyToRawTweet := range inReplyToRawTweets {
		inReplyTo := newTweet(&inReplyToRawTweet)
		if inReplyTo.text != "" {
			t := tweetsByInReplyToStatusID[inReplyToRawTweet.ID]
			d := dialog{tweets: []tweet{t, inReplyTo}}
			dialogs = append(dialogs, d)
		}
	}
	return dialogs
}

func (s *stream) processDialogs(ctx context.Context) {
	defer close(s.dialogs)

	resetTicker := time.NewTicker(statusLookupResetDuration)
	defer resetTicker.Stop()

	statusLookupCallCount := 0
	tweets := make([]tweet, 0, statusLookupMaxTweetCount)
	dialogs := make([]dialog, 0, statusLookupMaxTweetCount)
	unfinishedDialogs := make([]dialog, 0, statusLookupMaxTweetCount)

	for {
		select {
		case <-ctx.Done():
			return
		case <-resetTicker.C:
			statusLookupCallCount = 0
			log.Println("reset status lookup API limitation")
		case t := <-s.dialogTweets:
			tweets = append(tweets, t)
			if len(tweets) < statusLookupMaxTweetCount {
				continue
			}

			if statusLookupCallCount < statusLookupMaxCallCount {
				replyDialogs := s.makeDialogs(tweets[:statusLookupMaxTweetCount])
				tweets = tweets[:0]
				unfinishedDialogs = unfinishedDialogs[:0]

				for _, d := range replyDialogs {
					s.tweets <- d.lastTweet()
					for _, dlg := range dialogs {
						for _, t := range dlg.tweets {
							if t.id == d.firstTweet().id {
								dlg.tweets = append(dlg.tweets, d.lastTweet())
								d = dlg
							}
						}
					}
					if d.countUserIDs() != 2 {
						continue
					}
					if d.isFinished() {
						s.dialogs <- d
					} else {
						tweets = append(tweets, d.lastTweet())
						unfinishedDialogs = append(unfinishedDialogs, d)
					}
				}
				dialogs = append(dialogs[:0], unfinishedDialogs...)
				statusLookupCallCount++
				log.Println("call status lookup API", statusLookupCallCount)
			} else {
				tweets = tweets[:0]
			}
		}
	}
}

func (s *stream) processTweets(ctx context.Context) {
	defer close(s.dialogs)

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-s.stream.Messages:
			if rawTweet, ok := m.(*twitter.Tweet); ok {
				t := newTweet(rawTweet)
				if t.text == "" {
					continue
				}

				s.tweets <- t
				if t.hasInReplyTo() {
					select {
					case s.dialogTweets <- t:
					default:
						log.Println("dialogTweets channel is full")
					}
				}
			} else if err := m.(error); ok {
				log.Fatal(err)
			} else {
				log.Println(reflect.TypeOf(m).Name(), m)
			}
		}
	}
}

func newStream(ctx context.Context, language, consumeKey, consumeKeySecret, accessToken, accessTokenSecret string) (*stream, error) {
	config := oauth1.NewConfig(consumeKey, consumeKeySecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	streamParams := &twitter.StreamSampleParams{StallWarnings: twitter.Bool(true), Language: []string{language}}
	rawStream, err := client.Streams.Sample(streamParams)
	if err != nil {
		return nil, err
	}

	s := &stream{
		stream:       rawStream,
		client:       client,
		dialogTweets: make(chan tweet, 1000),

		dialogs: make(chan dialog),
		tweets:  make(chan tweet),
	}
	go s.processTweets(ctx)
	go s.processDialogs(ctx)
	return s, nil
}
