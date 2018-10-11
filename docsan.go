package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
	log "ibfd.org/docsan/log4u"
	"ibfd.org/docsan/render"
)

const maxFormParseMemorySizeBytes = 10 * 1024 * 1024

func main() {
	defer config.CloseLog()
	server := http.Server{Addr: ":" + config.GetPort()}
	log.Infof("docsan %s started on %s", version, server.Addr)
	http.HandleFunc("/", handle)
	server.ListenAndServe()
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		showForm(w)
	case "POST":
		process(w, r)
	default:
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

func process(w http.ResponseWriter, r *http.Request) {
	defer serverError(w, r)
	reader, err := getReader(r)
	if err == nil {
		htmlDoc, err := html.Parse(reader)
		if err == nil {
			document := render.Transform(htmlDoc, version)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			err = document.ToJSON(w)
		}
	}
	if err != nil {
		msg := fmt.Sprintf("failed to sanitize: %v", err)
		log.Errorf(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
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
		log.Debugf("key[%s] = %v\n", k, v)
	}
}

// ServerError maps errors to internal server errors.
func serverError(w http.ResponseWriter, rec *http.Request) {
	if r := recover(); r != nil {
		msg := fmt.Sprintf("%v", r)
		log.Error(msg)
		debug.PrintStack()
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
	}
}
