package main

import (
	"bufio"
	"fmt"
	"net"

	tls "github.com/refraction-networking/utls"
	"github.com/valyala/fasthttp"
)

var (
	roller *tls.Roller
	err    error
)

// FasthttpHTTPProxyRollerDialer uses uTLS roller to dial using a proxy
func FasthttpHTTPProxyRollerDialer(proxy string, server string) fasthttp.DialFunc {
	// var auth string
	// if strings.Contains(proxy, "@") {
	// 	split := strings.Split(proxy, "@")
	// 	auth = base64.StdEncoding.EncodeToString([]byte(split[0]))
	// 	proxy = split[1]
	// }
	return func(addr string) (net.Conn, error) {
		fmt.Println("Dialing using roller:")
		fmt.Println(proxy)
		fmt.Println(server)
		// Roller dial
		if roller == nil {
			roller, err = tls.NewRoller()
			if err != nil {
				return nil, err
			}
		}
		conn, err := roller.Dial("tcp", "client.tlsfingerprint.io:8443", "client.tlsfingerprint.io")
		if err != nil {
			return nil, err
		}
		// req := "CONNECT " + addr + " HTTP/1.1\r\n"
		// if auth != "" {
		// 	req += "Proxy-Authorization: Basic " + auth + "\r\n"
		// }
		// req += "\r\n"

		// if _, err := conn.Write([]byte(req)); err != nil {
		// 	return nil, err
		// }

		// res := fasthttp.AcquireResponse()
		// defer fasthttp.ReleaseResponse(res)

		// res.SkipBody = true

		// if err := res.Read(bufio.NewReader(conn)); err != nil {
		// 	conn.Close()
		// 	return nil, err
		// }
		// if res.Header.StatusCode() != 200 {
		// 	conn.Close()
		// 	return nil, fmt.Errorf("could not connect to proxy")
		// }
		return conn, nil
	}
}

// FasthttpHTTPRollerDialer uses uTLS roller to Dial
func FasthttpHTTPRollerDialer(addr string, server string) fasthttp.DialFunc {
	fmt.Println("Dialing using function here")
	return func(addr string) (net.Conn, error) {
		// Roller dial
		if roller == nil {
			roller, err = tls.NewRoller()
			if err != nil {
				return nil, err
			}
		}
		fmt.Println(addr)
		fmt.Println(server)
		conn, err := roller.Dial("tcp", addr, server)
		if err != nil {
			return nil, err
		}

		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)

		res.SkipBody = true

		if err := res.Read(bufio.NewReader(conn)); err != nil {
			conn.Close()
			return nil, err
		}
		if res.Header.StatusCode() != 200 {
			conn.Close()
			return nil, fmt.Errorf("could not connect to proxy")
		}
		return conn, nil
	}
}
