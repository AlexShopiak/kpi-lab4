package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/AlexShopiak/kpi-lab4/httptools"
	"github.com/AlexShopiak/kpi-lab4/signal"
)

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = []*Server{
		{url: "server1:8080"},
		{url: "server2:8080"},
		{url: "server3:8080"},
	}
	mutex sync.Mutex
)

type Server struct {
	healthy  bool
	connCnt int
	url string
}

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(server *Server) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), server.url), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(pool []*Server, index int, rw http.ResponseWriter, r *http.Request) error {
	if index == -1 {
		rw.WriteHeader(http.StatusServiceUnavailable)
		return fmt.Errorf("All servers are dead. No more healthy servers")
	}
	server := pool[index]
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = server.url
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = server.url

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		server.connCnt++
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", server.url)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", server.url, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

//Finds 1st healthy server
//Returns -1 if didnt find
//Finds server with less connections
func min(pool []*Server) int {
	index := -1
	minConn := -1

	for i, server := range serversPool {
		if server.healthy {
			index = i
			minConn = server.connCnt
		}
		
	}
	if index == -1 {
		return index
	}
	
	for i, server := range serversPool {
		if  server.healthy && server.connCnt < minConn  {
			index = i
			minConn = server.connCnt
		}   
	}
	return index
}

func main() {
	flag.Parse()

	// TODO: Використовуйте дані про стан сервреа, щоб підтримувати список тих серверів, яким можна відправляти ззапит.
	for _, server := range serversPool {
		server := server
		go func() {
			for range time.Tick(10 * time.Second) {
				mutex.Lock()
				server.healthy = health(server)
				log.Println(server.url,  server.connCnt, server.healthy)
				mutex.Unlock()
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		index := min(serversPool)
		mutex.Unlock()
		forward(serversPool, index, rw, r)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}
