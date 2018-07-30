package config

import (
	"flag"
	"os"
)

// DefaultPort defines the port to use if not defined
// by the environment or on the command line
const DefaultPort = "8080"

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
	allowedMetaNames := map[string]bool{
		"authorize_file":   true,
		"collection":       true,
		"pdf_chapter":      true,
		"print_version":    true,
		"titleframe_title": true,
		"word_chapter":     true}

	return func(metaName string) bool {
		_, present := allowedMetaNames[metaName]
		return present
	}

}
