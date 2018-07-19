package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"ibfd.org/docsan/san"
)

const maxFormParseMemorySizeBytes = 10 * 1024 * 1024

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		flag.Parse()
		port = flag.Arg(0)
		if port == "" {
			port = "8080"
		}
	}
	server := http.Server{Addr: ":" + port}
	log.Printf("docsan server started on %s", server.Addr)
	http.HandleFunc("/", handle)
	server.ListenAndServe()
}

func handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("method: %s: %s", r.Method, r.RequestURI)
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
						<input type="file" name="upload"><br>
						<input type="submit">
					</form>
				</body>
			</html>`
	w.Write([]byte(form))
}

func process(w http.ResponseWriter, r *http.Request) {
	for k, v := range r.Header {
		fmt.Printf("key[%s] value[%s]\n", k, v[0])
	}
	contentType := r.Header["Content-Type"]
	if contentType != nil && strings.HasPrefix(contentType[0], "multipart/form-data") {
		r.ParseMultipartForm(maxFormParseMemorySizeBytes)
		fileHeader := r.MultipartForm.File["upload"][0]
		file, err := fileHeader.Open()
		if err == nil {
			data, err := ioutil.ReadAll(file)
			if err == nil {
				sanitize(w, data)
				return
			}
		}
	} else {
		data, err := ioutil.ReadAll(r.Body)
		if err == nil {
			sanitize(w, data)
			return
		}
	}
	w.WriteHeader(500)
}

func sanitize(w http.ResponseWriter, data []byte) {
	log.Printf("file read succesfully %d bytes", len(data))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	san.Sanitize(w, string(data))
}

func fileSample() {
	filename := "evdeudir_2006_112.html"
	s := readFile(filename)
	err := san.Sanitize(os.Stdout, s)
	if err != nil {
		log.Fatalf("failed to sanitize file %s: %v", filename, err)
	}
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("fail to read file %s: %v", filename, err)
	}
	return string(data)
}
