package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

// DefaultPort defines the port to use if not defined
// by the environment or on the command line
const DefaultPort = "8080"
const configFilename = "config.json"

// Config defines the structure of the config.json file
type Config struct {
	MetaTags []string `json:"meta_tags"`
}

var allowedMetaNames map[string]bool

func init() {
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("fail to read file %s: %v", configFilename, err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("fail to unmarshal from file %s: %v", configFilename, err)
	}
	allowedMetaNames = make(map[string]bool, len(config.MetaTags))
	for _, metaName := range config.MetaTags {
		allowedMetaNames[metaName] = true
	}
	log.Printf("loaded configuration from %s", configFilename)
}

// GetPort returns the port to use for the Docsan service
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flag.Parse()
		port = flag.Arg(0)
		if port == "" {
			port = DefaultPort
		}
	}
	return port
}

// MetaNameAccept returns a function to filter meta tags
// TODO: Read allowed meta tags from config file.
func MetaNameAccept() func(string) bool {
	return func(metaName string) bool {
		_, present := allowedMetaNames[metaName]
		return present
	}
}
