package main

// RequestOpts : Request options received from node client
type RequestOpts struct {
	Method      string            `json:"method,omitempty"`
	URL         string            `json:"url,omitempty"`
	Proxy       string            `json:"proxy,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	HeaderOrder string            `json:"header_order,omitempty"`
	Body        string            `json:"body,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
	ParrotID    int               `json:"parrotId,omitempty"`
}

// RequestResp response to return
type RequestResp struct {
	Time       int                 `json:"timings,omitempty"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers,omitempty"`
}

// RequestResult result to return
type RequestResult struct {
	Error    string       `json:"error"`
	Response *RequestResp `json:"response,omitempty"`
	Body     string       `json:"body"`
}
