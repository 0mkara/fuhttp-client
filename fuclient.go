package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	tls "github.com/refraction-networking/utls"
	"github.com/valyala/fasthttp"
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
)

func fuclient(req *fasthttp.Request, res *fasthttp.Response, client *fasthttp.Client, SessionID string, ParrotID int, ch chan []byte) {
	// Finally do client request
	startTime := time.Now()
	timeout := time.Duration(20) * time.Second
	// Create Parrot map
	parrotmap[0] = parrots
	pm := parrotmap[0].([]tls.ClientHelloID)
	fucl := &fasthttp.Client{
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
		Dial:                          client.Dial,
	}
	// Load parrot
	// If ParrotID > 13 use custom parrots
	if ParrotID > 13 {
		hello, err := GetHelloCustom()
		if err != nil {
			ch <- []byte(`{"error":"Could not create custom hello spec."}`)
		}
		fucl.ClientHelloID = &tls.HelloCustom
		fucl.ClientHelloSpec = hello
	} else if ParrotID > -1 && ParrotID < 14 {
		fucl.ClientHelloID = &pm[ParrotID]
	} else {
		fucl.ClientHelloID = &pm[5]
	}
	// Else use predefined parrots
	// log.Println("fuclient: session id - ", SessionID)
	// fmt.Println(fucl.TLSConfig)
	if err := fucl.DoTimeout(req, res, timeout); err != nil {
		fmt.Println(err)
		ch <- []byte(`{"error":"` + err.Error() + `"}`)
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
					ch <- []byte(`{"error":"gzip read error"}`)
				}
				break
			case "br":
				bodyBytes, err = res.BodyUnbrotli()
				if err != nil {
					ch <- []byte(`{"error":"brotli read error"}`)
				}
				break
			case "deflate":
				bodyBytes, err = res.BodyInflate()
				if err != nil {
					ch <- []byte(`{"error":"deflate read error"}`)
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
	result.SessionID = SessionID
	result.Response = response
	result.Body = base64.StdEncoding.EncodeToString(bodyBytes)
	fb, err := json.Marshal(result)
	if err != nil {
		ch <- []byte(`{"error":"couldnt marshal json"}`)
	}
	log.Println(".............Final response.............")
	log.Println(string(fb))
	ch <- fb
}
