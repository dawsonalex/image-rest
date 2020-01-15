package imagebundle

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

// ImageController contains a number of functions for handling requests
// regarding images or image meta-data.
type ImageController struct {
	ContentDir string `json:"contentDir"` // The path to content. Defaults to user HOME
	images     []Image

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
		images:     make([]Image, 0),
	}
	return &controller
}

// SetLogger sets the logger on the referenced ImageController.
func (ic *ImageController) SetLogger(logger *log.Logger) {
	ic.logger = logger
}

// TODO: Move getContentType and loadImageMetaData to imageioutil package.

// getContentType returns the content-type of a file.
func getContentType(f *os.File) (string, error) {
	buffer := make([]byte, 512)

	_, err := f.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

// loadImageMetaData gets a slice of images stored at the path of contentDir.
func loadImageMetaData(contentDir string) ([]Image, error) {
	images := make([]Image, 0)

	// walk the path from contentDir load meta-data for each image.
	err := filepath.Walk(contentDir, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			file, err := os.Open(path)

			// if there is an error it is of type *PathError
			if err == nil {
				if contentType, err := getContentType(file); err == nil {
					if strings.Contains(contentType, "image/") {
						imageUUID, err := uuid.Parse(strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name())))
						// TODO: these errors should return nil to skip the image struct creation
						// and log a note that an image was skipped due to whatever error
						if err != nil {
							return err
						}

						imageURL, err := url.Parse(filepath.Join(path, fi.Name()))
						if err != nil {
							return err
						}

						image := Image{
							Name:        file.Name(),
							AbsoluteURL: imageURL,
							ID:          imageUUID,
						}
						images = append(images, image)
					}
				}
			}
		}
		return nil
	})

	return images, err
}

// HandleImageRequest returns image meta-data for the images in content directory.
func (ic *ImageController) HandleImageRequest() http.HandlerFunc {

	// When the handler is initialised, read the images from content dir.
	images, err := loadImageMetaData(ic.ContentDir)
	if err != nil {
		ic.logger.WithFields(log.Fields{
			"contentDirectory": ic.ContentDir,
			"err":              err,
		}).Error("An error occurred loading images")
	} else {
		ic.images = images
		ic.logger.WithFields(log.Fields{
			"image-count": len(ic.images),
		}).Info("Loaded images")
	}

	// return the HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(ic.images)
		if err != nil {
			ic.logger.WithFields(log.Fields{
				"err": err,
			}).Error("An error occurred marsahlling image data")

			w.WriteHeader(500)
			fmt.Fprintf(w, "Error occurred loading images")
			return
		}

		responsePayload := string(bytes[:len(bytes)])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, responsePayload)
	})
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

			uuidFileName := uuid.New().String() + filepath.Ext(part.FileName())
			fullImgFilePath := filepath.Join(ic.ContentDir, uuidFileName)
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
