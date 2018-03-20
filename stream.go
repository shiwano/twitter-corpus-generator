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

type dialog struct {
	tweets []tweet // order by tweeted at.
}

func (d *dialog) isFinished() bool {
	return d.lastTweet().inReplyToStatusID <= 0
}

func (d *dialog) firstTweet() tweet {
	return d.tweets[0]
}

func (d *dialog) lastTweet() tweet {
	return d.tweets[len(d.tweets)-1]
}

func (d *dialog) countUserIDs() int {
	userIDs := make(map[int64]struct{})
	for _, t := range d.tweets {
		userIDs[t.userID] = struct{}{}
	}
	return len(userIDs)
}

type tweet struct {
	id                int64
	userID            int64
	text              string
	inReplyToStatusID int64
}

func newTweet(t *twitter.Tweet) tweet {
	return tweet{
		id:                t.ID,
		userID:            t.User.ID,
		text:              normalizeText(t.Text),
		inReplyToStatusID: t.InReplyToStatusID,
	}
}

type tweetStream struct {
	stream           *twitter.Stream
	client           *twitter.Client
	tweetsForDialogs chan tweet

	dialogs chan dialog
	tweets  chan tweet
}

func (s *tweetStream) makeDialogs(tweets []tweet) []dialog {
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
	var replies []dialog
	for _, inReplyToRawTweet := range inReplyToRawTweets {
		inReplyTo := newTweet(&inReplyToRawTweet)
		if inReplyTo.text != "" {
			t := tweetsByInReplyToStatusID[inReplyToRawTweet.ID]
			d := dialog{tweets: []tweet{t, inReplyTo}}
			replies = append(replies, d)
		}
	}
	return replies
}

func (s *tweetStream) processDialogs(ctx context.Context) {
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
		case t := <-s.tweetsForDialogs:
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

func (s *tweetStream) processTweets(ctx context.Context) {
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
				if t.inReplyToStatusID > 0 {
					select {
					case s.tweetsForDialogs <- t:
					default:
						log.Println("tweetsForDialogs channel is full")
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

func newTweetStream(ctx context.Context, language, consumeKey, consumeKeySecret, accessToken, accessTokenSecret string) (*tweetStream, error) {
	config := oauth1.NewConfig(consumeKey, consumeKeySecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	streamParams := &twitter.StreamSampleParams{StallWarnings: twitter.Bool(true), Language: []string{language}}
	stream, err := client.Streams.Sample(streamParams)
	if err != nil {
		return nil, err
	}

	s := &tweetStream{
		stream:           stream,
		client:           client,
		tweetsForDialogs: make(chan tweet, 1000),

		dialogs: make(chan dialog),
		tweets:  make(chan tweet),
	}
	go s.processTweets(ctx)
	go s.processDialogs(ctx)
	return s, nil
}
