package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

// defaultPort defines the port to use if not defined
// by the environment or on the command line
const defaultPort = "8080"
const defaultConfigFilePath = "docsan.json"

// Config defines the structure of the config.json file
type Config struct {
	MetaTags []string `json:"meta_tags"`
}

var allowedMetaNames map[string]bool
var configFilePath string

func init() {
	flag.Parse()
	configFilePath = flag.Arg(0)
	if configFilePath == "" {
		configFilePath = defaultConfigFilePath
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("fail to read file %s: %v", configFilePath, err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("fail to unmarshal from file %s: %v", configFilePath, err)
	}
	allowedMetaNames = make(map[string]bool, len(config.MetaTags))
	for _, metaName := range config.MetaTags {
		allowedMetaNames[metaName] = true
	}
	log.Printf("loaded configuration from %s", configFilePath)
}

// GetPort returns the port to use for the Docsan service
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flag.Parse()
		port = flag.Arg(1)
		if port == "" {
			port = defaultPort
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
