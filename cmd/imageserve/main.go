package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/dawsonalex/image-rest/internal/bundles/imagebundle"
	"github.com/dawsonalex/image-rest/pkg/logware"
	"github.com/jessevdk/go-flags"
)

var options struct {
	// Path to config file.
	ConfigPath string `short:"c" long:"config" default:"nil" optional:"true" optional-value:"conf/config.json" description:"The path to the config file. Defaults to conf/config.json in the directory of the executable."`

	// Path to log file.
	LogPath string `short:"l" long:"log" default:"nil" optional:"true" optional-value:"log/img-rest.log" description:"The path to the log file. If the arguement is provided without a path, log/img-rest.log is used."`
}

func main() {
	// init logger
	logger := logware.NewStdOutLogger()

	_, err := flags.Parse(&options)
	optionsString := fmt.Sprintf("%+v", options)
	if err != nil {
		logger.WithFields(log.Fields{
			"error":        err.Error(),
			"default_args": optionsString,
		}).Warn("Unable to parse arguments. Using Defaults. (Value is nil for no value.)")
	} else {
		logger.WithFields(log.Fields{
			"args": optionsString,
		}).Info("Parsed args")
	}

	// Declaring image controller
	ic := imagebundle.NewImageController("", logger)

	// Set up routes
	router := http.NewServeMux()
	router.HandleFunc("/upload", ic.HandleUpload())

	// Set up server
	s := &http.Server{
		Addr:    ":8080",
		Handler: logware.LogRoute(logger, router),
	}
	logger.Println("up and running..")

	log.Fatal(s.ListenAndServe())
}
