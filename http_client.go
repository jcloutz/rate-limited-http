package rate_limited_http

import (
	"io"
	"net/http"
	"time"

	"go.uber.org/ratelimit"
)

// QueuedHttpClient will process http requests based on their provided priority
type QueuedHttpClient interface {
	Get(url string, priority Priority) (resp *http.Response, err error)
	Head(url string, priority Priority) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader, priority Priority) (resp *http.Response, err error)
	Put(url, contentType string, body io.Reader, priority Priority) (resp *http.Response, err error)
	Delete(url, contentType string, priority Priority) (resp *http.Response, err error)
	Do(req *http.Request, priority Priority) (*http.Response, error)
	Close() error
}

type ApiTask struct {
	Request *http.Request
	Result  chan *ApiTaskResult
}

type ApiTaskResult struct {
	Result *http.Response
	Err    error
}

var _ QueuedHttpClient = &httpClient{}

type httpClient struct {
	client      *http.Client
	rateLimiter ratelimit.Limiter
	queue       *priorityQueue
	tasks       chan *ApiTask
	quitChan    chan bool
}

type HttpClientOptions struct {
	HttpClient       *http.Client
	RateLimit        int
	WorkQueueMaxSize int
	WeightImmediate  float64
	WeightHigh       float64
	WeightMedium     float64
	WeightLow        float64
}

func NewHttpClient(optionFunc ...func(options *HttpClientOptions)) *httpClient {
	opts := HttpClientOptions{
		HttpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		RateLimit:        10,
		WorkQueueMaxSize: 500,
		WeightImmediate:  100,
		WeightHigh:       0.8,
		WeightMedium:     0.6,
		WeightLow:        0.4,
	}

	if optionFunc != nil {
		optionFunc[0](&opts)
	}

	client := httpClient{
		client:      opts.HttpClient,
		rateLimiter: ratelimit.New(opts.RateLimit),
		queue: NewPriorityQueue(func(pqOpts *PriorityQueueOptions) {
			pqOpts.WeightImmediate = opts.WeightImmediate
			pqOpts.WeightHigh = opts.WeightHigh
			pqOpts.WeightMedium = opts.WeightMedium
			pqOpts.WeightLow = opts.WeightLow
		}),
		tasks:    make(chan *ApiTask, opts.WorkQueueMaxSize),
		quitChan: make(chan bool, 1),
	}

	client.start()

	return &client
}

// Start will start a go routine to watch the queue and dispatch jobs
func (h *httpClient) start() {
	go func() {
		for {
			h.rateLimiter.Take()
			if !h.queue.Empty() {
				apiTask, _ := h.queue.Pop()
				// send to work channel
				h.tasks <- apiTask.Task()

			}

			select {
			case work := <-h.tasks:
				// fetch results
				res, err := h.client.Do(work.Request)

				// write to api task result channel to unblock caller
				work.Result <- &ApiTaskResult{
					Result: res,
					Err:    err,
				}
			case <-h.quitChan:

				return
			default:
				continue
			}
		}
	}()
}

// Close will shutdown the queue listener
func (h *httpClient) Close() error {
	go func() {
		h.quitChan <- true
	}()

	return nil
}

// enqueue will convert the uri and priority into an ApiAtask and queue it. A thunk
// is returned to the caller
func (h *httpClient) enqueue(request *http.Request, priority Priority) (*http.Response, error) {
	task := ApiTask{
		Request: request,
		Result:  make(chan *ApiTaskResult),
	}
	node := NewQItem(&task, priority)
	//wr := WorkRequest{Uri: uri, Result: make(chan io.Reader), Priority: priority}
	h.queue.Push(node)

	return h.loadThunk(&task)()
}

// loadThunk takes the result returned from the Result channel of the ApiTask
// and returns it to the calling function
func (h *httpClient) loadThunk(request *ApiTask) func() (*http.Response, error) {
	return func() (*http.Response, error) {
		res := <-request.Result

		return res.Result, res.Err
	}
}

func (h *httpClient) Get(url string, priority Priority) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return h.enqueue(req, priority)
}

func (h *httpClient) Head(url string, priority Priority) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	return h.enqueue(req, priority)
}

func (h *httpClient) Post(url, contentType string, body io.Reader, priority Priority) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	return h.enqueue(req, priority)
}

func (h *httpClient) Put(url, contentType string, body io.Reader, priority Priority) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	return h.enqueue(req, priority)
}

func (h *httpClient) Delete(url, contentType string, priority Priority) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return nil, err
	}

	return h.enqueue(req, priority)
}

func (h *httpClient) Do(req *http.Request, priority Priority) (resp *http.Response, err error) {
	return h.enqueue(req, priority)
}
