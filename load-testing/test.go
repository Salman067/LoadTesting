package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LoadTestClient represents a client for load testing.
type LoadTestClient struct {
	URL                   string
	NumRequests           int
	Timeout               time.Duration
	Results               chan time.Duration
	TotalTime             time.Duration
	AverageTime           time.Duration
	NumSuccessfulRequests int
	TotalErrors           int
	StatusMetrics         map[int]*StatusMetric
	Mutex                 sync.Mutex
	MinLatency            time.Duration
	MaxLatency            time.Duration
}

// StatusMetric represents metrics for a specific HTTP status code.
type StatusMetric struct {
	Count int
}

// NewLoadTestClient creates a new instance of LoadTestClient.
func NewLoadTestClient(url string, numRequests int, timeout time.Duration) *LoadTestClient {
	return &LoadTestClient{
		URL:           url,
		NumRequests:   numRequests,
		Timeout:       timeout,
		Results:       make(chan time.Duration, numRequests),
		StatusMetrics: make(map[int]*StatusMetric),
		MinLatency:    time.Duration(int(^uint(0) >> 1)), // Set MinLatency to max possible duration initially
		MaxLatency:    0,
	}
}

// Run executes the load test.
func (c *LoadTestClient) Run() {
	var wg sync.WaitGroup

	// Start timing
	start := time.Now()

	// Send concurrent HTTP requests
	for i := 0; i < c.NumRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startRequest := time.Now()
			client := http.Client{
				Timeout: c.Timeout,
			}
			resp, err := client.Get(c.URL)
			if err != nil {
				c.Mutex.Lock()
				c.TotalErrors++
				c.Mutex.Unlock()
				fmt.Println("Error:", err)
				return
			}
			defer resp.Body.Close()
			elapsed := time.Since(startRequest)
			c.Results <- elapsed

			// Update status code metrics
			c.Mutex.Lock()
			if c.StatusMetrics[resp.StatusCode] == nil {
				c.StatusMetrics[resp.StatusCode] = &StatusMetric{}
			}
			c.StatusMetrics[resp.StatusCode].Count++

			// Update min and max latency
			if elapsed < c.MinLatency {
				c.MinLatency = elapsed
			}
			if elapsed > c.MaxLatency {
				c.MaxLatency = elapsed
			}
			c.Mutex.Unlock()
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the results channel
	close(c.Results)

	// Calculate and store results
	c.TotalTime = time.Since(start)
	for r := range c.Results {
		c.NumSuccessfulRequests++
		c.AverageTime += r
	}
	c.AverageTime /= time.Duration(c.NumSuccessfulRequests)
}

func main() {
	var url string
	var numOfusers int
	var duration int

	// Define command-line flags
	flag.StringVar(&url, "url", "http://localhost:4000/dataFromDatabaseByParams?email=srvTYLL@QuHNUJJ.ru", "URL to test")
	flag.IntVar(&numOfusers, "users", 1000, "Number of concurrent requests")
	flag.IntVar(&duration, "duration", 10, "Request duration in seconds")
	flag.Parse()

	// Create a new load test client
	client := NewLoadTestClient(url, numOfusers, time.Duration(duration)*time.Second)

	// Run the load test
	client.Run()

	// Calculate metrics
	totalRequests := client.NumRequests
	totalErrors := client.TotalErrors
	totalTime := client.TotalTime
	avgLatency := client.AverageTime
	minLatency := client.MinLatency
	maxLatency := client.MaxLatency
	reqPerSec := float64(totalRequests) / totalTime.Seconds()
	statusMetrics := client.StatusMetrics

	// Print metrics
	fmt.Println("Total Number of Requests:", totalRequests)
	fmt.Println("Requests Per Second:", reqPerSec)
	fmt.Println("Average Latency:", avgLatency)
	fmt.Println("Min Latency:", minLatency)
	fmt.Println("Max Latency:", maxLatency)
	fmt.Println("Error Rate:", float64(totalErrors)/float64(totalRequests)*100, "%")
	fmt.Println("Status Code      Counts ")
	for status, metrics := range statusMetrics {
		fmt.Printf("%d               %d\n", status, metrics.Count)
	}
	fmt.Println("Total execution time", totalTime)
}
