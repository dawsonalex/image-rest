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

// RemoveHandler handlers requests to a delete a file from the server.
func RemoveHandler(dir string, logger *logrus.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodDelete) {
			logger.Errorf("Invalid HTTP method, got: %v", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		filenames := r.URL.Query()["name"]
		for _, file := range filenames {
			if filepath.Dir(file) != "." {
				logger.Debugf("invalid remove request for filename: %s", file)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, "name param should only contain a file name.")
				return
			}
			fullpath := filepath.Join(dir, filepath.Base(file))
			err := os.Remove(fullpath)
			if err != nil {
				if pErr, ok := err.(*os.PathError); ok && pErr == os.ErrNotExist {
					w.WriteHeader(http.StatusNotFound)
				}
				fmt.Fprint(w, "error removing files")
				w.WriteHeader(http.StatusInternalServerError)
				logger.Errorf("RemoveHandler(): error deleting file %s: %v", fullpath, err)
			}
		}
	})
}

// ImageHandler handles request for individual images.
func ImageHandler(dir string, logger *logrus.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodGet) {
			logger.Errorf("Invalid HTTP method, got: %v", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		imagenames, ok := r.URL.Query()["name"]
		if !ok {
			logger.Errorf("no name paramater provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		imagename := imagenames[0]
		if filepath.Dir(imagename) != "." {
			logger.Debugf("invalid request for filename: %s", imagename)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "name param should only contain a file name")
			return
		}
		imagepath := filepath.Join(dir, filepath.Base(imagename))
		http.ServeFile(w, r, imagepath)
	})
}

// UploadHandler handles requests to upload files to the server.
func UploadHandler(uploadDir string, logger *logrus.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodPost) {
			logger.Errorf("Invalid HTTP method, got: %v", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

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
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error occured during upload")
			return
		}

		// buffer to be used for reading bytes from files
		chunk := make([]byte, 4096)
		fileBytes := make([]byte, 0)

		contentTypeChecked := false
		for {
			// flag for tracking when the end of a part is reached.
			var uploaded bool

			if part, err = mr.NextPart(); err != nil {
				if err != io.EOF {
					logger.Errorf("Error occurred while fetching next part: %v", err)

					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Error occured during upload")
				} else {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintf(w, "")
				}
				return
			}

			// Read in the next chunk
			for !uploaded {
				// If we get an error reading the chunk EOF indicates chunk is done
				// any other error is a problem.
				if n, err = part.Read(chunk); err != nil {
					if err != io.EOF {
						logger.Errorf("Error occurred reading chunk: %v", err)

						w.WriteHeader(http.StatusInternalServerError)
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
						logger.Errorf("HandleUpload(): attempted to upload non-image file: %s", part.FileName())
						http.Error(w, "Request content is not an image", http.StatusUnsupportedMediaType)
						return
					}
					contentTypeChecked = true
				}

				// append this chunk to the whole file bytes in memory.
				fileBytes = append(fileBytes, chunk[:n]...)
			}

			// initially, write the file to temp dir,
			// then we'll copy the file to the watch dir in one go.
			// This avoids the imageservice getting multiple write
			// events when the file isn't fully loaded from the network.
			imgPath := filepath.Join(uploadDir, part.FileName())
			imgfile, err := os.Create(imgPath)
			if err != nil {
				logger.Errorf("Error occurred while creating image file: %v", err)

				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error occured during upload")
				return
			}
			defer imgfile.Close()

			// write the whole file to disc
			if _, err = imgfile.Write(fileBytes[:]); err != nil {
				logger.Errorf("Error occurred writing chunk to save file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error occured during upload")
				return
			}
			logger.Debugf("saved file %s", imgfile.Name())
		}
	})
}
