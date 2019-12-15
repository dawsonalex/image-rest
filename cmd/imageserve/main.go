package main

import (
	"fmt"
	"net/http"
	"os"

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

	ImgDir string `short:"d" long:"dir" required:"true" description:"The directory to serve images from."`
}

func main() {
	// init logger
	logger := logware.NewStdOutLogger()

	parseArgs(logger)

	// Declare image controller
	ic := imagebundle.NewImageController("", logger)

	// Set up routes
	router := http.NewServeMux()
	router.HandleFunc("/upload", ic.HandleUpload())

	// Set up server
	s := &http.Server{
		Addr:    ":8080",
		Handler: logware.LogRoute(logger, router),
	}
	logger.WithFields(log.Fields{
		"port": s.Addr,
	}).Info("Waiting for requests")

	log.Fatal(s.ListenAndServe())
}

// Parse command line arguments and log messages
func parseArgs(logger *log.Logger) {
	_, err := flags.NewParser(&options, flags.HelpFlag).Parse()
	if err != nil {
		flagError, _ := err.(*flags.Error)
		switch flagError.Type {
		// ErrHelp message contains the help page.
		case flags.ErrHelp:
			// Print the help message as-is and exit.
			logger.Print(flagError.Message)
			os.Exit(0)

		// Fatal errors parsing args, end program.
		case flags.ErrRequired,
			flags.ErrExpectedArgument:
			logger.Fatal(flagError.Message)

		// Non-fatal errors, warn and continue.
		default:
			logger.Warn("There was an error parsing arguements. Check your config is correct below.")
		}
	}

	optionsString := fmt.Sprintf("%+v", options)
	logger.WithFields(log.Fields{
		"args": optionsString,
	}).Info("Parsed args")
}
