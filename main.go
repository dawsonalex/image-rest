package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/dawsonalex/image-rest/server"

	"github.com/dawsonalex/image-rest/imageservice"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLogLevel = logrus.InfoLevel
	defaultPort     = 8080
)

var (
	logger   *logrus.Logger
	mountDir string
	logLevel logrus.Level
	port     *int
)

func init() {
	flag.StringVar(&mountDir, "dir", defaultDir(), "the path of the directory to watch")
	logLevelHelp := fmt.Sprintf("The level of logging to use, must be one of %v", logrus.AllLevels)
	var logLevelParam string
	flag.StringVar(&logLevelParam, "l", defaultLogLevel.String(), logLevelHelp)
	port = flag.Int("p", defaultPort, "The port to listen for API requests on.")
	flag.Parse()

	logger = initLogger(logLevelParam)
}

func main() {

	service := imageservice.New(logger)
	err := service.Watch(mountDir)
	if err != nil {
		logger.Fatalf("error watching dir %s: %v", mountDir, err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/list", server.FilesHandler(service, logger))
	router.HandleFunc("/upload", server.UploadHandler(mountDir, logger))
	router.HandleFunc("/remove", server.RemoveHandler(mountDir, logger))
	router.HandleFunc("/image", server.ImageHandler(mountDir, logger))

	// Set up server
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: logRoute(router),
	}
	// start the server log errors.
	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()
	logger.WithField("port", s.Addr).Info("Starting listening for requests")

	// await SIGINT from the OS, then cleanup.
	awaitInterrupt(func(done chan struct{}) {
		service.Stop()
		if err := s.Shutdown(context.Background()); err != nil {
			panic(err)
		}
		done <- struct{}{}
	})
}

func initLogger(level string) *logrus.Logger {
	logger := logrus.New()
	if logLevel, err := logrus.ParseLevel(level); err != nil {
		fmt.Printf("error during log init: %v\n", err)
		fmt.Println("using default log level 'info'")
		logger.SetLevel(defaultLogLevel)
	} else {
		fmt.Println("using log level: ", logLevel.String())
		logger.SetLevel(logLevel)
	}
	return logger
}

// logRoute is middleware for a server to log HTTP data about a
// request made to the server.
func logRoute(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log request details including reseponse.
		logger.WithFields(log.Fields{
			"protocol": r.Proto,
			"method":   r.Method,
			"route":    r.URL,
		}).Info("Request received")

		f.ServeHTTP(w, r)
	}
}

// awaitInterrupt blocks execution until it receives a SIGNINT
// from the OS.
func awaitInterrupt(onInterrupt func(chan struct{})) {
	done := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		defer signal.Stop(sigchan)

		<-sigchan
		logger.Println("received sigint")
		onInterrupt(done)
		logger.Println("shutdown.")
	}()
	<-done
}

func defaultDir() string {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Dir(ex)
}
