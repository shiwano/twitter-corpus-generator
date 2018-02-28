package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	t.Run(".normalizeText", func(t *testing.T) {
		text := normalizeText("@example_acount2 #exampleですよ http://example.com &gt;それなすぎる‥泣く。フレンド\nだったよね？？")
		assert.Equal(t, ">それなすぎる..泣く。フレンド だったよね??", text)
	})
}
