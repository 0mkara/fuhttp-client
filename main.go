package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	tls "github.com/refraction-networking/utls"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

// RequestOpts : Request options received from node client
type RequestOpts struct {
	Method      string            `json:"method,omitempty"`
	URL         string            `json:"url,omitempty"`
	Proxy       *string           `json:"proxy,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	HeaderOrder []string          `json:"header_order,omitempty"`
	Body        string            `json:"body"`
	Timeout     int               `json:"timeout"`
}

type RequestResp struct {
	Time       int                 `json:"timings,omitempty"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers,omitempty"`
}

type RequestResult struct {
	Error    string       `json:"error"`
	Response *RequestResp `json:"response,omitempty"`
	Body     string       `json:"body"`
}

var (
	sockAddr = "/tmp/fuhttp.sock"
	client   = &fasthttp.Client{
		Name:                          tls.HelloFirefox_56.Client,
		NoDefaultUserAgentHeader:      true,
		MaxConnsPerHost:               10000,
		ReadBufferSize:                4 * 4096, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               4 * 4096, // Same but for your response.
		ReadTimeout:                   time.Second,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this.
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
		ClientHelloID:                 &tls.HelloFirefox_56,
	}
)

func main() {
	log.Println("Starting new IPC server...")
	// Start IPC server
	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}
	s, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer s.Close()
	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := s.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go echoServer(conn)
		go reader(conn)
	}
}

func echoServer(c net.Conn) {
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())
}

func reader(c net.Conn) {
	buf := make([]byte, 1024*4)
	for {
		n, err := c.Read(buf[:])
		if err != nil {
			c.Write([]byte(err.Error()))
			return
		}
		fmt.Println("client request .............")
		fmt.Println(string(buf[0:n]))
		reqOpts := RequestOpts{}
		err = json.Unmarshal(buf[0:n], &reqOpts)
		if err != nil {
			c.Write([]byte(`{"error":"Request parsing failed: ` + err.Error() + `"}`))
			return
		}
		req := fasthttp.AcquireRequest()
		res := fasthttp.AcquireResponse()
		defer func() {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(res)
		}()
		// Load request parameters
		req.SetRequestURI(reqOpts.URL)
		req.Header.SetMethod(reqOpts.Method)
		// Load headers
		for h, i := range reqOpts.Headers {
			if h != "User-Agent" {
				req.Header.Add(h, i)
			}
		}
		// load json
		// load jar
		// load proxy
		if reqOpts.Proxy != nil {
			client.Dial = fasthttpproxy.FasthttpHTTPDialer(*reqOpts.Proxy)
		}
		// load body
		if reqOpts.Body != "" {
			req.AppendBodyString(reqOpts.Body)
		}
		fmt.Println(req)
		fmt.Println(string(req.URI().FullURI()))
		// finally do client request
		startTime := time.Now()
		timeout := time.Duration(60) * time.Second
		if err := client.DoTimeout(req, res, timeout); err != nil {
			c.Write([]byte(`{"error":"` + err.Error() + `"}`))
			c.Close()
		}
		// log.Println("Logging results.......................................")
		// // Body Reader
		// // fmt.Println("Print headers...............")
		// // fmt.Println(string(res.Header.Header()))
		// log.Println("Logging body.......................................")
		// fmt.Println(string(res.Body()))
		// log.Println("Logging body end.......................................")
		// fmt.Println(base64.StdEncoding.EncodeToString(res.Body()))
		// log.Println("Logging body EncodeToString.......................................")
		var bodyBytes []byte
		res.Header.VisitAll(func(key, value []byte) {
			if string(key) == "Content-Encoding" {
				log.Println("detecting encoding.......")
				log.Println(string(value))
				switch string(value) {
				case "gzip":
					bodyBytes, err = res.BodyGunzip()
					if err != nil {
						c.Write([]byte(`{"error":"gzip read error"}`))
					}
				case "br":
					bodyBytes, err = res.BodyUnbrotli()
					if err != nil {
						c.Write([]byte(`{"error":"brotli read error"}`))
					}
					break
				case "deflate":
					bodyBytes, err = res.BodyInflate()
					if err != nil {
						c.Write([]byte(`{"error":"brotli read error"}`))
					}
					break
				default:
					bodyBytes = res.Body()
				}
			}
		})
		if !(len(bodyBytes) > 0) {
			bodyBytes = res.Body()
		}
		response := &RequestResp{}
		response.Time = int(time.Since(startTime).Milliseconds())
		response.StatusCode = res.StatusCode()
		response.Headers = map[string][]string{}
		// Add all headers to response
		res.Header.VisitAll(func(key, value []byte) {
			response.Headers[string(key)] = append(response.Headers[string(key)], string(value))
		})

		result := &RequestResult{}
		result.Response = response
		result.Body = base64.StdEncoding.EncodeToString(bodyBytes)
		fb, err := json.Marshal(result)
		if err != nil {
			c.Write([]byte(`{"error":"couldnt marshal json"}`))
		}
		c.Write(fb)
		c.Close()
	}
}
