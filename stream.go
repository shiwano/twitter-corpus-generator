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

type reply struct {
	tweet     tweet
	inReplyTo tweet
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
	tweetsForReplies chan tweet

	replies chan reply
	tweets  chan tweet
}

func (s *tweetStream) makeReplies(tweets []tweet) []reply {
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
	var replies []reply
	for _, inReplyToRawTweet := range inReplyToRawTweets {
		inReplyTo := newTweet(&inReplyToRawTweet)
		if inReplyTo.text != "" {
			t := tweetsByInReplyToStatusID[inReplyToRawTweet.ID]
			replies = append(replies, reply{tweet: t, inReplyTo: inReplyTo})
		}
	}
	return replies
}

func (s *tweetStream) processReplies(ctx context.Context) {
	defer close(s.replies)

	resetTicker := time.NewTicker(statusLookupResetDuration)
	defer resetTicker.Stop()

	statusLookupCallCount := 0
	replyTweets := make([]tweet, 0, statusLookupMaxTweetCount)

	for {
		select {
		case <-ctx.Done():
			return
		case <-resetTicker.C:
			statusLookupCallCount = 0
			log.Println("reset status lookup API limitation")
		case t := <-s.tweetsForReplies:
			replyTweets = append(replyTweets, t)

			if len(replyTweets) == statusLookupMaxTweetCount {
				replyTweets = replyTweets[:0]

				if statusLookupCallCount < statusLookupMaxCallCount {
					replies := s.makeReplies(replyTweets)

					for _, r := range replies {
						s.tweets <- r.inReplyTo
						if r.tweet.userID != r.inReplyTo.userID {
							s.replies <- r
						}

						if r.inReplyTo.inReplyToStatusID > 0 {
							replyTweets = append(replyTweets, r.inReplyTo)
						}
					}
					statusLookupCallCount++
					log.Println("call status lookup API", statusLookupCallCount)
				}
			}
		}
	}
}

func (s *tweetStream) processTweets(ctx context.Context) {
	defer close(s.replies)

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
					case s.tweetsForReplies <- t:
					default:
						log.Println("tweetsForReplies channel is full")
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
		tweetsForReplies: make(chan tweet, 1000),

		replies: make(chan reply),
		tweets:  make(chan tweet),
	}
	go s.processTweets(ctx)
	go s.processReplies(ctx)
	return s, nil
}
