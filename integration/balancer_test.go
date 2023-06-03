package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
	. "gopkg.in/check.v1"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

type MySuite2 struct{}
var _ = Suite(&MySuite2{})
func Test(t *testing.T) { TestingT(t) }


func (s *MySuite2) TestBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	for i := 1; i <= 10; i++ {
		resp1, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil { c.Error(err)}
		header := resp1.Header.Get("lb-from")
		c.Assert(header, Equals, "server1:8080")
		c.Logf("response from [%s]", header)

		resp2, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil { c.Error(err)}
		header = resp2.Header.Get("lb-from")
		c.Assert(header, Equals, "server2:8080")
		c.Logf("response from [%s]", header)

		resp3, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil { c.Error(err)}
		header = resp3.Header.Get("lb-from")
		c.Assert(header, Equals, "server3:8080")
		c.Logf("response from [%s]", header)
	}
}

func (s *MySuite2) BenchmarkBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	for i := 0; i < c.N; i++ {
		_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			c.Error(err)
		}
	}
}
