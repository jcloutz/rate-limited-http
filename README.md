# Rate Limited & Queued Http Client

Proof of concept api client implementing a rate limited weighted priority queue
    
## Requirements
- Http requests must be rate limited to n requests to second
- Http requests must be able to have an associated priority
- Http requests must be FIFO within the scope of their priority level
- Http requests must be able to be prioritized as Immediate when it is imperative
that the user receives a quick response.

## Implementation
- __http_client/http_client.go__: 
    - Rate limited by requests/sec
    - Utilizes weighted priority queue
    
- __http_client/pq.go__:
    - Weighted priority queue
    - Priority levels are FIFO within their own scope
    - 4 levels of priority: Immediate, High, Medium, Low
    
## Usage
```go

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Create httpClient
	httpClient := http_client.NewHttpClient(func(opts *http_client.HttpClientOptions) {
		opts.RateLimit = 20 // requests per second
		opts.HttpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
		opts.WorkQueueMaxSize = 500 // channel size

		// [number priority items] * [weight] == weighted priority
		opts.WeightImmediate = 100
		opts.WeightHigh = 0.8
		opts.WeightMedium = 0.6
		opts.WeightLow = 0.4
	})
	defer httpClient.Close()

	go func() {
		if resp, err := httpClient.Get("https://jsonplaceholder.typicode.com/posts/", http_client.Low); err == nil {
			fmt.Println(fmt.Sprintf("completed low priority request status %d", resp.Status))
		}
	}()
	go func() {
		if resp, err := httpClient.Get("https://jsonplaceholder.typicode.com/posts/", http_client.Medium); err == nil {
			fmt.Println(fmt.Sprintf("completed medium priority request status %d", resp.Status))
		}
	}()
	go func() {
		if resp, err := httpClient.Get("https://jsonplaceholder.typicode.com/posts/", http_client.High); err == nil {
			fmt.Println(fmt.Sprintf("completed high priority request status %d", resp.Status))
		}
	}()
	go func() {
		if resp, err := httpClient.Get("https://jsonplaceholder.typicode.com/posts/", http_client.Immediate); err == nil {
			fmt.Println(fmt.Sprintf("completed immediate priority request status %d", resp.Status))
		}
	}()

	select {

	case sig := <-shutdown:
		log.Printf("main : starting shutdown... %v", sig)
		if err := httpClient.Close(); err != nil {
			log.Println(err)
		}
	}
}
```
