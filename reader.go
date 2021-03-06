package main

import (
	"bufio"
	"bytes"
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
	sessions = make(map[string]*fasthttp.Client)
	client   = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		EnableRawHeaders:              true,
		MaxConnsPerHost:               10000,
		ReadBufferSize:                4 * 4096, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               4 * 4096, // Same but for your response.
		ReadTimeout:                   time.Second * 10,
		WriteTimeout:                  time.Second * 10,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this.
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true, MaxVersion: 0},
	}
)

func reader(c net.Conn) {
	buf := make([]byte, 1024*1024*1)
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
		// TODO: session implementation
		// 1. check if have sessionid load client parameters
		if reqOpts.SessionID != "" && sessions[reqOpts.SessionID] != nil {
			fmt.Println("Loading session: ", reqOpts.SessionID)
			client = &fasthttp.Client{
				Name:                          sessions[reqOpts.SessionID].Name,
				NoDefaultUserAgentHeader:      sessions[reqOpts.SessionID].NoDefaultUserAgentHeader,
				EnableRawHeaders:              sessions[reqOpts.SessionID].EnableRawHeaders,
				MaxConnsPerHost:               sessions[reqOpts.SessionID].MaxConnsPerHost,
				ReadBufferSize:                sessions[reqOpts.SessionID].ReadBufferSize,
				WriteBufferSize:               sessions[reqOpts.SessionID].WriteBufferSize,
				ReadTimeout:                   sessions[reqOpts.SessionID].ReadTimeout,
				WriteTimeout:                  sessions[reqOpts.SessionID].WriteTimeout,
				MaxIdleConnDuration:           sessions[reqOpts.SessionID].MaxIdleConnDuration,
				DisableHeaderNamesNormalizing: sessions[reqOpts.SessionID].DisableHeaderNamesNormalizing,
				TLSConfig:                     sessions[reqOpts.SessionID].TLSConfig.Clone(),
				Dial:                          sessions[reqOpts.SessionID].Dial,
			}
		}
		// 2. create new session variables or load from existing session
		if reqOpts.SessionID != "" || sessions[reqOpts.SessionID] == nil {
			fmt.Println("Creating session: ", reqOpts.SessionID)
			client.Name = reqOpts.Name
			// Load proxy
			if reqOpts.Proxy != "" {
				client.Dial = fasthttpproxy.FasthttpHTTPDialer(reqOpts.Proxy)
			}
			sessions[reqOpts.SessionID] = &fasthttp.Client{
				Name:                          client.Name,
				NoDefaultUserAgentHeader:      client.NoDefaultUserAgentHeader,
				EnableRawHeaders:              client.EnableRawHeaders,
				MaxConnsPerHost:               client.MaxConnsPerHost,
				ReadBufferSize:                client.ReadBufferSize,
				WriteBufferSize:               client.WriteBufferSize,
				ReadTimeout:                   client.ReadTimeout,
				WriteTimeout:                  client.WriteTimeout,
				MaxIdleConnDuration:           client.MaxIdleConnDuration,
				DisableHeaderNamesNormalizing: client.DisableHeaderNamesNormalizing,
				TLSConfig:                     client.TLSConfig.Clone(),
				Dial:                          client.Dial,
			}
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
		ch := make(chan []byte, 4)
		go fuclient(req, res, client, reqOpts.SessionID, reqOpts.ParrotID, ch)
		c.Write([]byte(<-ch))
		close(ch)
		c.Close()
	}
}
