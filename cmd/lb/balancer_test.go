package main

import (
	"testing"
	. "gopkg.in/check.v1"
)


type MySuite struct{}
var _ = Suite(&MySuite{})

func TestBalancer(t *testing.T) { TestingT(t) }

func (s *MySuite) TestMin(c *C) {
	serversPool := []*Server{
		{url: "Server1", connCnt: 0, healthy: false},
		{url: "Server2", connCnt: 0, healthy: false},
		{url: "Server3", connCnt: 0, healthy: false},
	}
	c.Assert(min(serversPool), Equals, -1)

	serversPool = []*Server{
		{url: "Server1", connCnt: 0, healthy: true},
		{url: "Server2", connCnt: 0, healthy: false},
		{url: "Server3", connCnt: 0, healthy: false},
	}
	c.Assert(min(serversPool), Equals, 0)

	serversPool = []*Server{
		{url: "Server1", connCnt: 0, healthy: false},
		{url: "Server2", connCnt: 0, healthy: false},
		{url: "Server3", connCnt: 0, healthy: true},
	}
	c.Assert(min(serversPool), Equals, 2)

	serversPool = []*Server{
		{url: "Server1", connCnt: 20, healthy: true},
		{url: "Server2", connCnt: 10, healthy: true},
		{url: "Server3", connCnt: 30, healthy: true},
	}
	c.Assert(min(serversPool), Equals, 1)

	serversPool = []*Server{
		{url: "Server1", connCnt: 20, healthy: true},
		{url: "Server2", connCnt: 10, healthy: true},
		{url: "Server3", connCnt: 10, healthy: true},
	}
	c.Assert(min(serversPool), Equals, 1)
  }