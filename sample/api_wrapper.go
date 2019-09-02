package main

import (
	"encoding/json"
	"fmt"
	"strings"

	http_client "github.com/jcloutz/rate-limited-http"
)

type PostApiWrapper struct {
	client http_client.QueuedHttpClient
}

func NewPostApiWrapper(httpClient http_client.QueuedHttpClient) *PostApiWrapper {
	return &PostApiWrapper{client: httpClient}
}

// Fetch post will call jsonplaceholder and retrieve the specified post
func (api *PostApiWrapper) FetchPost(id int, priority http_client.Priority) (*Post, error) {
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

func (api *PostApiWrapper) CreatePost(title string, body string, priority http_client.Priority) (*Post, error) {
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

func (api *PostApiWrapper) UpdatePost(id int, title string, body string, priority http_client.Priority) (*Post, error) {
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

func (api *PostApiWrapper) DeletePost(id int, priority http_client.Priority) (*Post, error) {
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
