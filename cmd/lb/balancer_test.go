package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"
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

func (s *MySuite) TestHealth(c *C) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", "http://api.mybiz.com/health",
		httpmock.NewStringResponder(200, ""))
	server := &Server{url: "api.mybiz.com"}

	// do stuff that makes a request to articles
	healthy := health(server)

	c.Assert(healthy, Equals, true)
}
