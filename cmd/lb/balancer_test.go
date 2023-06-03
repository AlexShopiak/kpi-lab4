package main

import (
	"testing"

	"github.com/jarcoal/httpmock"
	. "gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
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

func (s *MySuite) TestScheme(c *C) {
	*https = true
	c.Assert(scheme(), Equals, "https")

	*https = false
	c.Assert(scheme(), Equals, "http")
}

func (s *MySuite) TestForward(c *C) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", "http://server1:8080/data",
		httpmock.NewStringResponder(200, "OK"))
	pool := []*Server{{url: "server1:8080", healthy: true}}
	rr := httptest.NewRecorder()

	// do stuff that makes a request to articles
	req, err := http.NewRequest("GET", "/data", nil)
	c.Assert(err, Equals, nil)

	err = forward(pool, 0, rr, req)
	c.Assert(err, Equals, nil)

	err = forward(pool, -1, rr, req)
	c.Assert(err, ErrorMatches, "All servers are dead. No more healthy servers")

	pool[0].url = "server2:8080"
	err = forward(pool, 0, rr, req)
	c.Assert(err, ErrorMatches, "Get \"http://server2:8080/data\": no responder found")
}
