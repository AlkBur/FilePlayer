package main

import (
	"log"
	"net/http"
	"time"
	"os"
)

// Headers
const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderVary            = "Vary"
	gzipScheme            = "gzip"
	HeaderContentEncoding = "Content-Encoding"
)

var l *log.Logger

func init() {
	l = log.New(os.Stdout, "DEBUG: ", log.Lshortfile|log.LstdFlags)
}

func logging(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer trace(r)()
		//w.Header().Add(HeaderVary, HeaderAcceptEncoding)
		//if strings.Contains(r.Header.Get(HeaderAcceptEncoding), gzipScheme) {
		//	w.Header().Add(HeaderContentEncoding, gzipScheme)
		//	wg, err := gzip.NewWriterLevel(w, 2)
		//	if err != nil {
		//		return
		//	}
		//}
		//defer func() {
		//	wg.Close()
		//}()
		handler.ServeHTTP(w, r)
	})
}

func trace(r *http.Request) func() {
	start := time.Now()
	return func() {
		l.Printf("Method: %s, url: %s, completed in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

func logDebug(name string, arg interface{})  {
	if debug {
		l.Printf(name, GetString(arg))
	}
}
