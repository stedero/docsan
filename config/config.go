package config

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"ibfd.org/docsan/log4u"
)

// defaultPort defines the port to use if not defined
// by the environment or on the command line
const defaultPort = "8080"
const defaultConfigFilePath = "docsan.json"
const defaultLogLevel = "DEBUG"

// LogDef defines logging configuration
type LogDef struct {
	Filename string `json:"filename"`
	Level    string `json:"level"`
}

// Config defines the structure of the config.json file
type Config struct {
	Logging  LogDef   `json:"logging"`
	MetaTags []string `json:"meta_tags"`
}

var allowedMetaNames map[string]bool
var configFilePath string
var logFile *os.File

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
	logFile = configureLogging(&config.Logging)
	allowedMetaNames = make(map[string]bool, len(config.MetaTags))
	for _, metaName := range config.MetaTags {
		allowedMetaNames[metaName] = true
	}
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

// CloseLog closes the log file.
func CloseLog() {
	logFile.Close()
}

func configureLogging(logConfig *LogDef) *os.File {
	var logFile *os.File
	var err error
	if logConfig == nil || logConfig.Filename == "" {
		log4u.SetLevel(defaultLogLevel)
	} else {
		logFile, err = os.Create(logConfig.Filename)
		if err != nil {
			log.Fatalf("failed to create file %s: %v", logConfig.Filename, err)
		}
		logger := io.MultiWriter(os.Stderr, logFile)
		log4u.SetLevel(logConfig.Level)
		log4u.SetOutput(logger)
	}
	return logFile
}

// MetaNameAccept returns a function to filter meta tags
func MetaNameAccept() func(string) bool {
	return func(metaName string) bool {
		_, present := allowedMetaNames[metaName]
		return present
	}
}
