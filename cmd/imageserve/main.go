package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/dawsonalex/image-rest/internal/bundles/imagebundle"
	"github.com/dawsonalex/image-rest/pkg/logware"
	"github.com/jessevdk/go-flags"
)

var options struct {
	// Path to log file.
	LogPath string `short:"l" long:"log" default:"log/img-rest.log" description:"The path to the log file."`

	// Path to directory to serve
	ImgDir string `short:"d" long:"dir" default:"." description:"The path to the directory to serve images from."`
}

func main() {
	// init logger
	logger := logware.NewStdOutLogger()

	parseArgs(logger)

	// Declare image controller
	ic := &imagebundle.ImageController{
		ContentDir: options.ImgDir,
	}
	ic.SetLogger(logger)

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

	// If the imagedir path is default, set it to dir of
	// the executable that started the process.
	if options.ImgDir == "." {
		ex, err := os.Executable()
		if err != nil {
			logger.Panicln(err)
		}
		exPath := filepath.Dir(ex)
		options.ImgDir = exPath
	}

	optionsString := fmt.Sprintf("%+v", options)
	logger.WithFields(log.Fields{
		"args": optionsString,
	}).Info("Parsed args")
}
