package main

import (
	"github.com/dghubble/go-twitter/twitter"
)

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

func (t *tweet) hasInReplyTo() bool {
	return t.inReplyToStatusID > 0
}

type dialog struct {
	tweets []tweet // order by tweeted at.
}

func (d *dialog) isFinished() bool {
	t := d.lastTweet()
	return !t.hasInReplyTo()
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
