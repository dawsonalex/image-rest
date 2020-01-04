package imagebundle

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ImageController contains a number of functions for handling requests
// regarding images or image meta-data.
type ImageController struct {
	ContentDir string `json:"contentDir"` // The path to content. Defaults to user HOME

	logger *log.Logger
}

func defaultContentDir() string {
	contentDir, err := os.UserHomeDir()
	if err != nil {
		contentDir = os.TempDir()
	}

	return contentDir
}

// NewImageController returns a new ImageController initialised with default values.
func NewImageController(logger *log.Logger) *ImageController {
	controller := ImageController{
		ContentDir: defaultContentDir(),
		logger:     logger,
	}
	return &controller
}

// SetLogger sets the logger on the referenced ImageController.
func (ic *ImageController) SetLogger(logger *log.Logger) {
	ic.logger = logger
}

// HandleUpload saves images from a multipart form submission to disk.
func (ic *ImageController) HandleUpload() http.HandlerFunc {

	// check here that the directory is writable. If not, log and error.
	if fileMode, err := os.Stat(ic.ContentDir); err != nil {
		ic.logger.WithFields(log.Fields{
			"error":      err,
			"contentDir": ic.ContentDir,
		}).Error("There is an error reading from the content directory:")
	} else if !fileMode.IsDir() {
		ic.logger.WithFields(log.Fields{
			"contentDir": ic.ContentDir,
		}).Error("The contentDir provided is not a directory")
	}

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
			// imgfile: filehandler for the temporary file
			// filesize: how many bytes where written to the imgfile
			// uploaded: boolean to flip when the end of a part is reached
			var imgfile *os.File
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

			//imgfile, err = ioutil.TempFile(os.TempDir(), "example-upload-*.tmp")
			fullImgFilePath := filepath.Join(ic.ContentDir, part.FileName())
			imgfile, err = os.Create(fullImgFilePath)
			if err != nil {
				ic.logger.WithFields(log.Fields{
					"err": err.Error(),
				}).Error("Error occurred while creating image file")

				w.WriteHeader(500)
				fmt.Fprintf(w, "Error occured during upload")
				return
			}
			// defer tempfile close and removal.
			defer imgfile.Close()

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

				if n, err = imgfile.Write(chunk[:n]); err != nil {
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
			baseFileName := filepath.Base(imgfile.Name())
			ic.logger.WithFields(log.Fields{
				"filename": baseFileName,
				"fullpath": fullImgFilePath,
				"size":     filesize,
			}).Info("Image saved")
		}
	})
}
