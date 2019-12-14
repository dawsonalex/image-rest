package imagebundle

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type imageBundleConfig struct {
	ContentDir string `json:"contentDir"` // The path to content. Defaults to user HOME
}

// ImageController contains a number of functions for handling requests
// regarding images or image meta-data.
type ImageController struct {
	config *imageBundleConfig
	logger *log.Logger
}

func defaultImageBundleConfig() *imageBundleConfig {
	bundleConfig := new(imageBundleConfig)

	userHome, err := os.UserHomeDir()
	if err != nil {
		bundleConfig.ContentDir = os.TempDir()
	} else {
		bundleConfig.ContentDir = userHome
	}

	return bundleConfig
}

func jsonImageBundleConfig(filePath string) *imageBundleConfig {
	return defaultImageBundleConfig()
}

// NewImageController returns a new ImageController initialised
// with a the config at configPath, or a default config.
func NewImageController(configPath string, logger *log.Logger) *ImageController {
	var bundleConfig *imageBundleConfig
	if len(configPath) > 0 {
		// read in image controller config from a JSON file
		bundleConfig = jsonImageBundleConfig(configPath)
	} else {
		bundleConfig = defaultImageBundleConfig()
	}

	controller := ImageController{
		config: bundleConfig,
		logger: logger,
	}
	return &controller
}

// HandleUpload sadf
func (ic *ImageController) HandleUpload() http.HandlerFunc {

	// contents of this function taken from https://www.reddit.com/r/golang/comments/apf6l5/multiple_files_upload_using_gos_standard_library/
	// Read multipart form data from multiple files and write them to disc.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// define some variables used throughout the function
		// n: for keeping track of bytes read and written
		// err: for storing errors that need checking
		var n int
		var err error

		// define pointers for the multipart reader and its parts
		var mr *multipart.Reader
		var part *multipart.Part

		if mr, err = r.MultipartReader(); err != nil {
			ic.logger.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Error opening multipart reader")

			w.WriteHeader(500)
			fmt.Fprintf(w, "Error occured during upload")
			return
		}

		// buffer to be used for reading bytes from files
		chunk := make([]byte, 4096)

		// continue looping through all parts, *multipart.Reader.NextPart() will
		// return an End of File when all parts have been read.
		for {
			// variables used in this loop only
			// tempfile: filehandler for the temporary file
			// filesize: how many bytes where written to the tempfile
			// uploaded: boolean to flip when the end of a part is reached
			var tempfile *os.File
			var filesize int
			var uploaded bool

			if part, err = mr.NextPart(); err != nil {
				if err != io.EOF {
					ic.logger.WithFields(log.Fields{
						"err": err.Error(),
					}).Error("Error occurred while fetching next part")

					w.WriteHeader(500)
					fmt.Fprintf(w, "Error occured during upload")
				} else {
					w.WriteHeader(200)
					fmt.Fprintf(w, "Upload complete")
				}
				return
			}

			tempfile, err = ioutil.TempFile(os.TempDir(), "example-upload-*.tmp")
			if err != nil {
				ic.logger.WithFields(log.Fields{
					"err": err.Error(),
				}).Error("Error occurred while creating temp file")

				w.WriteHeader(500)
				fmt.Fprintf(w, "Error occured during upload")
				return
			}
			// defer tempfile close and removal.
			defer tempfile.Close()
			defer os.Remove(tempfile.Name())

			// continue reading until the whole file is upload or an error is reached
			for !uploaded {
				if n, err = part.Read(chunk); err != nil {
					if err != io.EOF {
						ic.logger.WithFields(log.Fields{
							"err": err.Error(),
						}).Error("Error occurred reading chunk")

						w.WriteHeader(500)
						fmt.Fprintf(w, "Error occured during upload")
						return
					}
					uploaded = true
				}

				if n, err = tempfile.Write(chunk[:n]); err != nil {
					ic.logger.WithFields(log.Fields{
						"err": err.Error(),
					}).Error("Error occurred writing chunk to save file")

					w.WriteHeader(500)
					fmt.Fprintf(w, "Error occured during upload")
					return
				}
				filesize += n
			}

			// Only print the name of the file, not the full filepath.
			baseFileName := filepath.Base(tempfile.Name())
			ic.logger.WithFields(log.Fields{
				"filename": baseFileName,
				"size":     filesize,
			}).Info("Image saved")
		}
	})

}
