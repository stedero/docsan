package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
	"ibfd.org/docsan/render"
)

const maxFormParseMemorySizeBytes = 10 * 1024 * 1024

func main() {
	server := http.Server{Addr: ":" + config.GetPort()}
	log.Printf("docsan %s started on %s", version, server.Addr)
	http.HandleFunc("/", handler(config.MetaAccept()))
	server.ListenAndServe()
}

func handler(metaAccept func(string) bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("method: %s: %s", r.Method, r.RequestURI)
		switch r.Method {
		case "GET":
			showForm(w)
		case "POST":
			process(metaAccept, w, r)
		default:
		}
	}
}

func showForm(w http.ResponseWriter) {
	form := `<html>
				<head>
					<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
					<title>DocSan - HTML document sanitizer</title>
				</head>
				<body>
					<h1>DocSan - HTML document sanitizer</h1>
					<form action="/" method="post" enctype="multipart/form-data">
						<input type="file" name="upload"><br><br>
						<input type="submit">
					</form>
				</body>
			</html>`
	w.Write([]byte(form))
}

func process(metaAccept func(string) bool, w http.ResponseWriter, r *http.Request) {
	logHeaders(r)
	reader, err := getReader(r)
	if err == nil {
		htmlDoc, err := html.Parse(reader)
		if err == nil {
			jsonData, err := render.ToJSON(htmlDoc, version)
			if err == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(200)
				w.Write(jsonData)
			}
		}
	}
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("failed to sanitize: %v", err)))
	}
}

func getReader(r *http.Request) (io.Reader, error) {
	contentType := r.Header["Content-Type"]
	if contentType != nil && strings.HasPrefix(contentType[0], "multipart/form-data") {
		r.ParseMultipartForm(maxFormParseMemorySizeBytes)
		fileHeader := r.MultipartForm.File["upload"][0]
		return fileHeader.Open()
	}
	return r.Body, nil
}

func logHeaders(r *http.Request) {
	for k, v := range r.Header {
		fmt.Printf("key[%s] = %v\n", k, v)
	}
}
