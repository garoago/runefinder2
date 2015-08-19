package main

import (
	"gopkg.in/check.v1"
	"strings"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type MySuite struct{}

var _ = check.Suite(&MySuite{})

func (s *MySuite) TestFindRunes(c *check.C) {
	index := map[string]runeSlice{
		"REGISTERED": runeSlice{0xAE},
		"BLACK":      runeSlice{0x265A, 0x265B, 0x265C, 0x265D, 0x265E, 0x265F, 0x2B24 /* not chess */},
		"CHESS":      runeSlice{0x265A, 0x265B, 0x265C, 0x265D, 0x265E, 0x265F, 0x2654 /* not black */},
	}

	tests := map[string]runeSlice{
		"registered":  runeSlice{0xAE},
		"nonesuch":    runeSlice{},
		"chess black": runeSlice{0x265A, 0x265B, 0x265C, 0x265D, 0x265E, 0x265F},
	}
	for query, found := range tests {
		words := strings.Split(query, " ")
		c.Check(findRunes(words, index), check.DeepEquals, found)
	}
}
