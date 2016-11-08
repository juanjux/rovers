package bing

import (
	"testing"

	"github.com/src-d/rovers/core"
	"github.com/src-d/rovers/utils/websearch"
	. "gopkg.in/check.v1"
)

type BingSuite struct {
	bing  websearch.Searcher
	query string
}

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&BingSuite{
	query: "\"powered by cgit\"||\"generated by cgit\"||\"Commits per author per week\"",
})

func (s *BingSuite) SetUpTest(c *C) {
	s.bing = New(core.Config.Bing.Key)
}

func (s *BingSuite) TestBing_Search(c *C) {
	result, err := s.bing.Search(s.query)

	c.Assert(err, IsNil)
	c.Assert(result, NotNil)
	c.Assert(len(result) > 200, Equals, true)
}

func (s *BingSuite) TestBing_BadKey(c *C) {
	bing := New("BAD_KEY")
	result, err := bing.Search(s.query)

	c.Assert(err, Equals, errInvalidKey)
	c.Assert(result, IsNil)
}
