package server

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dawsonalex/image-rest/imageservice"
	"github.com/sirupsen/logrus"
)

// FilesHandler returns a http.HandlerFunc that accepts requests for an image stores
// file list.
func FilesHandler(store *imageservice.Service, logger *logrus.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.Errorf("Invalid HTTP method, got: %v", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		images := store.Files()
		imageResponse := make([]imageservice.Image, 0)
		for _, image := range images {
			imageResponse = append(imageResponse, image)
		}
		imageResponse = sortFiles(imageResponse)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(imageResponse)
	})
}

// softFiles sorts image
func sortFiles(images []imageservice.Image) []imageservice.Image {
	sort.SliceStable(images, func(i, j int) bool {
		return images[i].Name < images[j].Name
	})
	return images
}

// UploadHandler handles requests to upload files to the server.
func UploadHandler(uploadDir string, logger *logrus.Logger) http.HandlerFunc {
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
			logger.Errorf("Error opening multipart reader: %v", err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error occured during upload")
			return
		}

		// buffer to be used for reading bytes from files
		chunk := make([]byte, 4096)

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
					logger.Errorf("Error occurred while fetching next part: %v", err)

					w.WriteHeader(500)
					fmt.Fprintf(w, "Error occured during upload")
				} else {
					w.WriteHeader(200)
					fmt.Fprintf(w, "Upload complete")
				}
				return
			}

			imgPath := filepath.Join(uploadDir, part.FileName())
			imgfile, err = os.Create(imgPath)
			if err != nil {
				logger.Errorf("Error occurred while creating image file: %v", err)

				w.WriteHeader(500)
				fmt.Fprintf(w, "Error occured during upload")
				return
			}
			defer imgfile.Close()

			contentTypeChecked := false
			// Read in the next chunk
			for !uploaded {
				// If we get an error reading the chunk EOF indicates chunk is done
				// any other error is a problem.
				if n, err = part.Read(chunk); err != nil {
					if err != io.EOF {
						logger.Errorf("Error occurred reading chunk: %v", err)

						w.WriteHeader(500)
						fmt.Fprintf(w, "Error occured during upload")
						return
					}
					uploaded = true
				}

				// If we haven't tested the content type of the actual file,
				// do it now. Stop the upload if the file isn't an image.
				if !contentTypeChecked {
					contentType := http.DetectContentType(chunk)
					logger.Debugf("UploadHandler(): got image of content type %s", contentType)
					isImage := strings.Contains(contentType, "image/")
					if !isImage {
						logger.Errorf("HandleUpload(): attempted to upload non-image file - %s", imgfile.Name())
						http.Error(w, "Request content is not an image", http.StatusBadRequest)
						return
					}
					contentTypeChecked = true
				}

				// Write the bytes we read from the part into the chunk to
				// our file.
				if n, err = imgfile.Write(chunk[:n]); err != nil {
					logger.Errorf("Error occurred writing chunk to save file: %v", err)

					w.WriteHeader(500)
					fmt.Fprintf(w, "Error occured during upload")
					return
				}
				filesize += n
			}

			logger.WithFields(logrus.Fields{
				"filename": imgfile.Name(),
				"size":     filesize,
			}).Info("image saved")
		}
	})
}
