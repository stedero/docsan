package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"ibfd.org/docsan/render"
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
						<input type="file" name="upload"><br><br>
						<input type="submit">
					</form>
				</body>
			</html>`
	w.Write([]byte(form))
}

func process(w http.ResponseWriter, r *http.Request) {
	var err error
	for k, v := range r.Header {
		fmt.Printf("key[%s] value[%s]\n", k, v[0])
	}
	accept := strings.Split(r.Header["Accept"][0], ",")
	fmt.Printf("Accept: %v", accept)
	if accept[0] == "application/json" {
		err = render.ToJSON(w, r.Body)
	} else {
		contentType := r.Header["Content-Type"]
		if contentType != nil && strings.HasPrefix(contentType[0], "multipart/form-data") {
			r.ParseMultipartForm(maxFormParseMemorySizeBytes)
			fileHeader := r.MultipartForm.File["upload"][0]
			file, err := fileHeader.Open()
			if err == nil {
				err = render.ToJSON(w, file)
			}
		} else {
			err = render.ToJSON(w, r.Body)
		}
	}
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("failed to sanitize: %v", err)))
	}
}
