package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/valyala/fasthttp"
)

func fuclient(c net.Conn, req *fasthttp.Request, res *fasthttp.Response, client *fasthttp.Client) {
	// Finally do client request
	startTime := time.Now()
	timeout := time.Duration(10) * time.Second
	fucl := &fasthttp.Client{
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
		ClientHelloSpec:               client.ClientHelloSpec,
		ClientHelloID:                 client.ClientHelloID,
		Dial:                          client.Dial,
	}
	if err := fucl.DoTimeout(req, res, timeout); err != nil {
		log.Println("Error in DoTimeout")
		fmt.Println(err)
		c.Write([]byte(`{"error":"` + err.Error() + `"}`))
		c.Close()
		return
	}

	var bodyBytes []byte
	var err error
	res.Header.VisitAll(func(key, value []byte) {
		if string(key) == "Content-Encoding" {
			switch string(value) {
			case "gzip":
				bodyBytes, err = res.BodyGunzip()
				if err != nil {
					c.Write([]byte(`{"error":"gzip read error"}`))
					c.Close()
				}
			case "br":
				bodyBytes, err = res.BodyUnbrotli()
				if err != nil {
					c.Write([]byte(`{"error":"brotli read error"}`))
					c.Close()
				}
				break
			case "deflate":
				bodyBytes, err = res.BodyInflate()
				if err != nil {
					c.Write([]byte(`{"error":"brotli read error"}`))
					c.Close()
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
