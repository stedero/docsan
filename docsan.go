package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
	log "ibfd.org/docsan/log4u"
	"ibfd.org/docsan/render"
)

var noFileError error

func main() {
	defer config.CloseLog()
	noFileError = errors.New("no file provided")
	server := http.Server{Addr: ":" + config.GetPort()}
	log.Infof("%s started on %s", appName(), server.Addr)
	df := render.NewDocumentFactory(appName())
	http.HandleFunc("/", handler(df))
	server.ListenAndServe()
}

func handler(df *render.DocumentFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			showForm(w)
		case "POST":
			process(df, w, r)
		default:
		}
	}
}

func showForm(w http.ResponseWriter) {
	setServer(w)
	form := `<html>
				<head>
					<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
					<title>` + appName() + ` - HTML document sanitizer</title>
				</head>
				<body>
					<h1>` + appName() + ` - HTML document sanitizer</h1>
					<form action="/" method="post" enctype="multipart/form-data">
						<input type="file" name="upload"><br><br>
						<input type="submit">
					</form>
				</body>
			</html>`
	w.Write([]byte(form))
}

func process(df *render.DocumentFactory, w http.ResponseWriter, r *http.Request) {
	defer serverError(w, r)
	reader, err := getReader(r)
	if err == nil {
		total := timer()
		htmlDoc, err := html.Parse(reader)
		if err == nil {
			document := df.Transform(htmlDoc)
			setServer(w)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(200)
			err = document.ToJSON(w)
			if err != nil {
				log.Errorf("%s: transformation of %s failed: %v", r.Host, document.DocID, err)
			} else if document.DocID != "" {
				log.Debugf("%s: transforming %s took %s", r.Host, document.DocID, total())
			}
		}
	}
	if err != nil {
		if err == noFileError {
			writeError(w, 400, err.Error())
		} else {
			writeError(w, 500, fmt.Sprintf("failed to sanitize: %v", err))
		}
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	if status >= 500 {
		log.Errorf(msg)
	}
	setServer(w)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func getReader(r *http.Request) (io.Reader, error) {
	contentType := r.Header["Content-Type"]
	if contentType != nil && strings.HasPrefix(contentType[0], "multipart/form-data") {
		r.ParseMultipartForm(1 << 20)
		fileHeaders := r.MultipartForm.File["upload"]
		if fileHeaders == nil {
			return nil, noFileError
		}
		fileHeader := fileHeaders[0]
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
		logStackDump()
		setServer(w)
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(msg))
	}
}

func logStackDump() {
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, true)
	stackdump := string(buf[0:stackSize])
	entries := strings.Split(stackdump, "\n")
	for _, entry := range entries {
		log.Errorf("\t%s", entry)
	}
}

func timer() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}

func setServer(w http.ResponseWriter) {
	w.Header().Set("Server", appName())
}

func appName() string {
	return "Docsan/" + version
}
