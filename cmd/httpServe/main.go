package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gohxs/prettylog"
	"github.com/stdiopt/httpserve"
)

func main() {
	prettylog.Global()

	var proxyTo string
	var flagMdCSS string

	flag.StringVar(&flagMdCSS, "md-css", "", "add a css file while rendering markdown")
	flag.StringVar(&proxyTo, "proxy", "", "do not serve files only creates a reverse proxy")
	flag.Parse()

	log.Println("V:", httpserve.Version)

	var handler http.Handler
	var logger *log.Logger
	if len(proxyTo) != 0 {
		logger = prettylog.New("proxy")
		log.Println("Proxy to:", proxyTo)
		u, err := url.Parse(proxyTo)
		if err != nil {
			log.Fatal(err)
		}
		handler = httputil.NewSingleHostReverseProxy(u)
	} else {
		logger = prettylog.New("files")
		h, err := httpserve.New(httpserve.Options{
			FlagMdCSS: flagMdCSS,
		})
		if err != nil {
			log.Fatal(err)
		}
		handler = h
	}
	mw := httpserve.Middlewares{httpserve.Logger(logger)}

	handler = mw.Apply(handler)

	port := 8080
	for {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Println("Err opening", port, err)
			port++
			log.Println("Trying port", port)
			continue
		}

		log.Printf("Listening at:")

		addrW := bytes.NewBuffer(nil)
		fmt.Fprintf(addrW, "    http://localhost:%d\n", port)
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			log.Fatal("err:", err)
		}
		for _, a := range addrs {
			astr := a.String()
			if strings.HasPrefix(astr, "192.168") ||
				strings.HasPrefix(astr, "10") {
				a := strings.Split(astr, "/")[0]
				fmt.Fprintf(addrW, "    http://%s:%d\n", a, port)
			}
		}
		log.Println(addrW.String())

		if err := http.Serve(listener, handler); err != nil {
			log.Println("Err serving", err)
		}

	}
}
