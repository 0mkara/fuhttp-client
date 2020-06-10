package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	tls "github.com/refraction-networking/utls"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	parrotmap = make(map[int]interface{})
	parrots   = []tls.ClientHelloID{
		tls.HelloFirefox_Auto,
		tls.HelloFirefox_55,
		tls.HelloFirefox_56,
		tls.HelloFirefox_63,
		tls.HelloFirefox_65,
		tls.HelloChrome_Auto,
		tls.HelloChrome_58,
		tls.HelloChrome_62,
		tls.HelloChrome_70,
		tls.HelloChrome_72,
		tls.HelloChrome_83,
		tls.HelloIOS_Auto,
		tls.HelloIOS_11_1,
		tls.HelloIOS_12_1,
	}
	client = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		EnableRawHeaders:              true,
		MaxConnsPerHost:               10000,
		ReadBufferSize:                4 * 4096, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               4 * 4096, // Same but for your response.
		ReadTimeout:                   time.Second,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this.
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
	}
)

func reader(c net.Conn) {
	buf := make([]byte, 1024*1024*1)
	parrotmap[0] = parrots
	pm := parrotmap[0].([]tls.ClientHelloID)
	for {
		n, err := c.Read(buf[:])
		if err != nil {
			c.Write([]byte(err.Error()))
			return
		}
		log.Println(".............Received request.............")
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
		// Load parrot
		if reqOpts.ParrotID > -1 {
			client.ClientHelloID = &pm[reqOpts.ParrotID]
		} else {
			client.ClientHelloID = &pm[5]
		}
		// Load headers in order if present
		if reqOpts.HeaderOrder != "" {
			r := bytes.NewBufferString(reqOpts.HeaderOrder)
			br := bufio.NewReader(r)
			if err := req.Header.Read(br); err != nil {
				log.Fatalf("Unexpected error: %s", err)
			}
		}
		// Load headers unordered
		if reqOpts.HeaderOrder == "" {
			for h, i := range reqOpts.Headers {
				if h != "Host" && h != "" {
					req.Header.Set(h, i)
				}
			}
		}
		// Load proxy
		if reqOpts.Proxy != nil {
			client.Dial = fasthttpproxy.FasthttpHTTPDialer(*reqOpts.Proxy)
		}
		// Load request URL
		req.SetRequestURI(reqOpts.URL)
		// Load request method
		if reqOpts.Method != "" {
			req.Header.SetMethod(reqOpts.Method)
		}
		// Load body
		if reqOpts.Body != "" {
			req.AppendBodyString(reqOpts.Body)
		}
		// Request parsing ends above
		// -------------------------------------------------------------------------------------------------------------------------------------
		// Finally do client request
		startTime := time.Now()
		timeout := time.Duration(60) * time.Second
		if err := client.DoTimeout(req, res, timeout); err != nil {
			c.Write([]byte(`{"error":"` + err.Error() + `"}`))
			c.Close()
			return
		}
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
		log.Println(".............Final response.............")
		log.Println(string(fb))
		c.Write(fb)
		c.Close()
	}
}
