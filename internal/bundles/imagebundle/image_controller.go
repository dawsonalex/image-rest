package imagebundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dawsonalex/image-rest/internal/imagelibrary"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrImageNotFound indicates that the requested image is not in the library.
	ErrImageNotFound = errors.New("Image not in library")
)

// Describes a JSON response for a request for a single image.
type imageResponse struct {
	Name   string    `json:"name"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	ID     uuid.UUID `json:"id"`
}

// Describes a JSON response for a request for a group of images.
type imageLibraryResponse struct {
	Count   int             `json:"count"`
	Library []imageResponse `json:"library"`
}

// Describes an error response.
type errorResponse struct {
	Msg string `json:"msg"`
}

// ImageController contains a number of functions for handling requests
// regarding images or image meta-data.
type ImageController struct {
	ContentDir  string `json:"contentDir"` // The path to content. Defaults to user HOME
	images      map[uuid.UUID]*imagelibrary.Image
	libraryInit bool // true if the image library has been loaded, otherwise false.

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
		images:     make(map[uuid.UUID]*imagelibrary.Image),
	}
	return &controller
}

// SetLogger sets the logger on the referenced ImageController.
func (ic *ImageController) SetLogger(logger *log.Logger) {
	ic.logger = logger
}

// InitLibrary creates an image library pointing to dir.
func (ic *ImageController) InitLibrary(dir string) {
	// When the handler is initialised, read the images from content dir.
	images, err := imagelibrary.FromDir(ic.ContentDir)

	if err != nil {
		ic.images = make(map[uuid.UUID]*imagelibrary.Image, 0)
		ic.logger.WithFields(log.Fields{
			"contentDirectory": ic.ContentDir,
			"err":              err,
		}).Error("An error occurred loading images")
	} else {
		ic.images = make(map[uuid.UUID]*imagelibrary.Image, len(images))
		for _, image := range images {
			ic.images[uuid.New()] = image
		}
		ic.logger.WithFields(log.Fields{
			"image-count": len(ic.images),
		}).Info("Loaded images")
	}

	ic.libraryInit = true
}

// HandleLibraryRequest returns image meta-data for the images in content directory.
func (ic *ImageController) HandleLibraryRequest() http.HandlerFunc {
	if !ic.libraryInit {
		ic.InitLibrary(ic.ContentDir)
	}

	// return the HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract imagelibrary.Image values into our JSON response structs.
		responseLibrary := make([]imageResponse, 0)
		for imgKey, img := range ic.images {
			image := imageResponse{
				Name:   img.Name,
				Width:  img.Width,
				Height: img.Height,
				ID:     imgKey,
			}

			responseLibrary = append(responseLibrary, image)
		}

		response := imageLibraryResponse{
			Count:   len(ic.images),
			Library: responseLibrary,
		}

		// Marshal the JSON response struct and return
		bytes, err := json.Marshal(response)
		if err != nil {
			ic.logger.WithFields(log.Fields{
				"err": err,
			}).Error("An error occurred marsahlling image data")

			w.WriteHeader(500)
			fmt.Fprintf(w, "Error occurred loading images")
			return
		}

		// Set the response body to our marshalled JSON
		// and send.
		responseBody := string(bytes[:len(bytes)])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, responseBody)
	})
}

// HandleImageRequest returns a handler func that manages requests for
// an image or list of images from the image library.
func (ic *ImageController) HandleImageRequest() http.HandlerFunc {
	if !ic.libraryInit {
		ic.InitLibrary(ic.ContentDir)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestQuery := r.URL.Query()
		imageID := requestQuery.Get("id")

		// If imageID is present, fetch the
		if len(imageID) > 0 {
			image, err := ic.ImageByID(imageID)

			// If the image isn't present, return some error JSON.
			if err != nil {
				ic.logger.WithFields(log.Fields{
					"value": imageID,
					"err":   err,
				}).Error("Error getting image")

				errorResponse := errorResponse{
					Msg: err.Error(),
				}

				errorResponseJSON, _ := json.Marshal(errorResponse)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(404)
				fmt.Fprintf(w, string(errorResponseJSON))
				return
			}

			// If the image is found serve it.
			ic.logger.WithFields(log.Fields{
				"id": imageID,
			}).Info("Responding with image")
			http.ServeFile(w, r, image.AbsoluteURL)
			return
		}

	})
}

// ImageByID returns an image from the library with matching ID.
func (ic *ImageController) ImageByID(id string) (*imagelibrary.Image, error) {
	imageUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrImageNotFound
	}

	if image, found := ic.images[imageUUID]; found {
		return image, nil
	}
	return nil, ErrImageNotFound
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
