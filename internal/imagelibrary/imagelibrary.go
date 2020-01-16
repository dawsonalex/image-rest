package imagelibrary

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type (
	// Image is an implementation of an image file on disk
	Image struct {
		Name        string    `json:"name"`
		AbsoluteURL *url.URL  `json:"url"`
		Width       int       `json:"width"`
		Height      int       `json:"height"`
		UUID        uuid.UUID `json:"id"`
	}
)

// FromDir loads the library images from a directory on disk.
func FromDir(imageDir string) *ImageLibrary {
	images, err := imageSliceFromDir(imageDir)

	// make images as an empty slice if an error occurred
	if err != nil {
		images = make([]Image, 0)
	}

	library := ImageLibrary{
		images,
		log.New(),
	}
	return &library
}

// ImageLibrary loads and stores meta-data on images from disk in memory.
type ImageLibrary struct {
	Images []Image
	logger *log.Logger
}

// SetLogger sets the logger object on the ImageLibrary.
func (il *ImageLibrary) SetLogger(logger *log.Logger) {
	il.logger = logger
}

// TODO: us ioutil.ReadDir to iterate the directory and get files. Then file.Stat and DecodeConfig to get other details.

func imageSliceFromDir(dir string) ([]Image, error) {
	images := make([]Image, 0)
	createImage := func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			file, err := os.Open(path)
			defer file.Close()

			if err == nil {
				if isImage, _ := isImageContentType(file); isImage {

					images = append(images, image)
				}
			}
		}

		return nil
	}

	err := filepath.Walk(dir, createImage)
	return images, err
}

func imageFromFile(file *os.File) Image {
	imageUUID := uuid.New()
	imageURL, _ := url.Parse(filepath.Join(path, file.Name()))

	image := Image{
		UUID:        imageUUID,
		AbsoluteURL: imageURL,
		Name:        fi.Name(),
	}
}

// isImageContentType returns true file the file contents is an image MIME-type.
// Otherwise returns false.
func isImageContentType(f *os.File) (bool, error) {
	buffer := make([]byte, 512)

	_, err := f.Read(buffer)
	if err != nil {
		return false, err
	}

	contentType := http.DetectContentType(buffer)
	isImage := strings.Contains(contentType, "image/")
	return isImage, nil
}
