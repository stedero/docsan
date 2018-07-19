package main

import (
	"io/ioutil"
	"log"
	"os"

	"ibfd.org/docsan/san"
)

func main() {
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
