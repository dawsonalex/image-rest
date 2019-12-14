package logware

import (
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var logger *log.Logger

// responseLog wraps http.ResponseWriter and can be used to
// inspect data about a server response after a call to
// ServeHttp has been made.
type responseLog struct {
	http.ResponseWriter
	status int
}

// Write implements the http.ResponseWriter.Write func.
func (r *responseLog) Write(p []byte) (int, error) {
	return r.ResponseWriter.Write(p)
}

// WriteHeader implemenets the http.ResponseWriter.WriteHeader func to
// store the repsonse status in r.
func (r *responseLog) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// NewMultiLogger creates a new Logger. The filepath variable is a path to the file
// that should be created or appended to
func NewMultiLogger(filePath string) *log.Logger {
	logFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Cannnot use logfile %v:", err)
	}

	writer := io.MultiWriter(os.Stderr, logFile)

	logger := log.New()
	logger.Out = writer
	logger.Formatter = &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	}
	return logger
}

// NewStdOutLogger returns a logrus instance that writes in colour to stdout.
func NewStdOutLogger() *log.Logger {
	logger := log.New()
	logger.Out = os.Stdout
	logger.Formatter = &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	}
	return logger
}

// LogRoute writes out information regarding requests in the format:
// [HTTP Proto] [HTTP Method] [Endpoint] -> [Response Status]
func LogRoute(logger *log.Logger, f http.Handler) http.HandlerFunc {
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
