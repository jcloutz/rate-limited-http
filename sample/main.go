package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/ratelimit"

	http_client "github.com/jcloutz/rate-limited-http"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Create httpClient
	httpClient := http_client.NewHttpClient(func(opts *http_client.HttpClientOptions) {
		opts.RateLimit = 2 // requests per second
	})
	defer httpClient.Close()

	client := NewPostApiWrapper(httpClient)

	// seed queue at a rate of 5 requests per second,
	rl := ratelimit.New(5)
	go func() {
		for i := 0; i < 15; i++ {
			rl.Take()
			priority := http_client.Priority(rand.Intn(5-1) + 1)
			entityId := rand.Intn(20-1) + 1
			action := rand.Intn(5 - 1)

			go func() {

				switch action {
				case 0:
					if _, err := client.FetchPost(entityId, priority); err != nil {
						fmt.Println(err)
					}
				case 1:
					if _, err := client.CreatePost("foo", "bar", priority); err != nil {
						fmt.Println(err)
					}
				case 2:
					if _, err := client.UpdatePost(1, "foo", "bar", priority); err != nil {
						fmt.Println(err)
					}
				case 3:
					if _, err := client.DeletePost(1, priority); err != nil {
						fmt.Println(err)
					}

				}
			}()

		}
	}()

	select {

	case sig := <-shutdown:
		log.Printf("main : starting shutdown... %v", sig)
		if err := httpClient.Close(); err != nil {
			log.Println(err)
		}

		switch {
		case sig == syscall.SIGSTOP:
			log.Fatal("main : Integrity issue caused shutdown")
		}
	}
}

type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}
