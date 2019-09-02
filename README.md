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
    
