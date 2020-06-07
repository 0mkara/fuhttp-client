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

	"github.com/valyala/fasthttp"
)

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

// func createOrderedHeader(h map[string]string) {
// 	m := orderedmap.NewOrderedMap()
// 	for i := range headerOredr {
// 		m.set(headerOredr[i], value)
// 	}
// }

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
		// ----------------------------------------------------------------
		// Test exceptions
		// c.Write([]byte(`{"error":"could not connect to proxy"}`))
		// c.Close()
		// ----------------------------------------------------------------
		reqOpts := RequestOpts{}
		err = json.Unmarshal(buf[0:n], &reqOpts)
		if err != nil {
			c.Write([]byte(`{"error":"Request parsing failed: ` + err.Error() + `"}`))
			return
		}
		log.Println("request opts parsed........")
		fmt.Println(reqOpts)
		req := fasthttp.AcquireRequest()
		res := fasthttp.AcquireResponse()
		defer func() {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(res)
		}()
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
				if h != "User-Agent" && h != "Host" && h != "" {
					req.Header.Set(h, i)
				}
			}
		}
		// Load cookies
		// for c, i := range reqOpts.Cookies {
		// 	req.Header.SetCookie(c, i)
		// }
		// Load proxy
		// if reqOpts.Proxy != nil {
		// 	client.Dial = fasthttpproxy.FasthttpHTTPDialer(*reqOpts.Proxy)
		// }
		// Load request URL
		req.SetRequestURI(reqOpts.URL)
		// Load request method
		req.Header.SetMethod(reqOpts.Method)
		// Load body
		if reqOpts.Body != "" {
			req.AppendBodyString(reqOpts.Body)
		}
		log.Println("................................................................................")
		fmt.Println("Request for ", string(req.URI().FullURI()))
		// fmt.Println()
		req.Header.VisitAllInOrder(func(key, value []byte) {
			// log.Println("visit request headers in order")
			fmt.Printf("%s:%s\n", string(key), string(value))
		})
		log.Println("................................................................................")
		req.Header.VisitAll(func(key, value []byte) {
			// log.Println("visit request headers in order")
			fmt.Printf("%s:%s\n", string(key), string(value))
		})
		// finally do client request
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
		log.Println("Final result:")
		log.Println(string(fb))
		c.Write(fb)
		c.Close()
	}
}
