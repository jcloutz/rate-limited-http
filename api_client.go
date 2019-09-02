package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jcloutz/rate-limited-http/http_client"
)

type ApiClient struct {
	client http_client.RateLimitedHttpClient
}

func NewApiClient(httpClient http_client.RateLimitedHttpClient) *ApiClient {
	return &ApiClient{client: httpClient}
}

// Fetch post will call jsonplaceholder and retrieve the specified post
func (api *ApiClient) FetchPost(id int, priority http_client.Priority) (*Post, error) {
	uri := fmt.Sprintf(`https://jsonplaceholder.typicode.com/posts/%d`, id)

	fmt.Println(fmt.Sprintf(`Queuing post id=%d, priority=%d`, id, priority))
	resp, err := api.client.Get(uri, priority)
	if err != nil {
		return nil, err
	}
	post := Post{}

	// decode post from io.Reader returned from load
	err = json.NewDecoder(resp.Body).Decode(&post)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf(`[P%d]GET: Loaded post id=%d, data=%v`, priority, post.ID, post))
	return &post, nil
}

func (api *ApiClient) CreatePost(title string, body string, priority http_client.Priority) (*Post, error) {
	uri := `https://jsonplaceholder.typicode.com/posts`

	fmt.Println(fmt.Sprintf(`Queuing create post title=%s, priority=%d`, title, priority))
	jsonData := fmt.Sprintf(`{"title": "%s", "body": "%s", "userId": %d}`, title, body, 1)
	resp, err := api.client.Post(uri, "application/json", strings.NewReader(jsonData), priority)
	if err != nil {
		return nil, err
	}

	post := Post{}
	err = json.NewDecoder(resp.Body).Decode(&post)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf(`[P%d]POST: Created post id=%d, data=%v`, priority, post.ID, post))
	return &post, nil
}

func (api *ApiClient) UpdatePost(id int, title string, body string, priority http_client.Priority) (*Post, error) {
	uri := fmt.Sprintf(`https://jsonplaceholder.typicode.com/posts/%d`, id)

	fmt.Println(fmt.Sprintf(`Queuing update post id=%d, priority=%d`, id, priority))
	jsonData := fmt.Sprintf(`{"id": %d, "title": "%s", "body": "%s", "userId": %d}`, id, title, body, 1)
	resp, err := api.client.Put(uri, "application/json", strings.NewReader(jsonData), priority)
	if err != nil {
		return nil, err
	}

	post := Post{}
	err = json.NewDecoder(resp.Body).Decode(&post)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf(`[P%d]PUT: Updated post id=%d, data=%v`, priority, post.ID, post))
	return &post, nil
}

func (api *ApiClient) DeletePost(id int, priority http_client.Priority) (*Post, error) {
	uri := fmt.Sprintf(`https://jsonplaceholder.typicode.com/posts/%d`, id)

	fmt.Println(fmt.Sprintf(`Queuing delete post id=%d, priority=%d`, id, priority))
	resp, err := api.client.Delete(uri, "application/json", priority)
	if err != nil {
		return nil, err
	}

	post := Post{}
	err = json.NewDecoder(resp.Body).Decode(&post)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf(`[P%d]DELETE: Deleted post id=%d, data=%v`, priority, post.ID, post))
	return &post, nil
}
