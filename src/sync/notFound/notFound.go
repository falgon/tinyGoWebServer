package notFound

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
)

type hook404 struct {
	http.ResponseWriter
	R         *http.Request
	handle404 func(w http.ResponseWriter, r *http.Request) bool
}

func (h *hook404) WriteHeader(code int) {
	h.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf8")
	h.ResponseWriter.WriteHeader(code)
	if 404 == code && h.handle404(h.ResponseWriter, h.R) {
		panic(h)
	}
}

func Handle404(handler http.Handler, handle404 func(w http.ResponseWriter, r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hook := &hook404 {
			ResponseWriter: w,
			R:              r,
			handle404:      handle404,
		}

		defer func() {
			if p := recover(); p != nil {
				if p == hook {
					return
				}
				panic(p)
			}
		}()
		handler.ServeHTTP(hook, r)
	})
}

func TryRead404(w http.ResponseWriter, fname string) {
	var f *os.File
	defer f.Close()
	if f, err := os.Open(fname); err != nil {
		io.WriteString(w, "<!DOCTYPE html><html lang=\"ja\"><head><meta charset=\"UTF-8\"><title>404 not found</title></heed><body><h1>404 not Found</h1></body></html>")
	} else {
		sc := bufio.NewScanner(f)
		for sc.Scan() && sc.Err() == nil {
			fmt.Fprintf(w, sc.Text())
		}
	}
}
