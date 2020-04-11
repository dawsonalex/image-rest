package imagelibrary

import (
	_ "image/gif"  // Register gif image decoding
	_ "image/jpeg" // Register jpeg image decoding
	_ "image/png"  // Register PNG image decoding
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// ImageContentType is the content type prefix for images.
	ImageContentType string = "image/"
)

type (
	// Image is an implementation of an image file on disk
	Image struct {
		Name        string `json:"name"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		AbsoluteURL string
	}

	// ImageCollection is a map of images, by a UUID.
	ImageCollection map[uuid.UUID]*Image

	// Library is a collection of images, mapped by IDs.
	Library struct {
		images ImageCollection
		logger *logrus.Logger
	}
)

// FromDir constructs a new Library from the images at
// the path dir.
func FromDir(dir string, logger *logrus.Logger, onImageLoad func(error)) (*Library, error) {
	images := make(ImageCollection, 0)
	library := &Library{
		images,
		logger,
	}

	go func() {
		err := library.loadDir(dir)
		onImageLoad(err)
	}()

	return library, nil
}

func (lib *Library) loadDir(dir string) error {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	images := make(ImageCollection)
	for _, fileInfo := range fileInfos {
		absoluteFilePath := filepath.Join(dir, fileInfo.Name())
		if file, err := os.Open(absoluteFilePath); err == nil {
			defer file.Close()

			if fileIsImage, err := fileIsImage(file); !fileIsImage || err != nil {
				continue
			}

			image := &Image{
				AbsoluteURL: absoluteFilePath,
				Name:        file.Name(),
			}
			lib.AddFile(file)
		} else {
			return err
		}
	}

	return nil
}

// AddFile adds an image to the librarys image collection.
func (lib *Library) AddFile(file *os.File) error {
	if fileIsImage, err := fileIsImage(file); !fileIsImage {
		return ErrFileNotImage{
			file.Name(),
		}
	} else if err != nil {
		return err
	}

	// TODO: add logic to add file to library and disc
	return nil
}

// FromDir2 returns a slice of images at the path dir.
func FromDir2(dir string) ([]*Image, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return make([]*Image, 0), err
	}

	imageSlice := make([]*Image, 0)
	for _, img := range files {
		fullFilePath := filepath.Join(dir, img.Name())
		log.Printf("Loading %s", fullFilePath)
		if reader, err := os.Open(fullFilePath); err == nil {
			defer reader.Close()
			if isImage, err := fileIsImage(reader); !isImage {
				log.Printf("file %s is not an image", reader.Name())
				continue
			} else if err != nil {
				log.Printf("error: %v", err)
			}

			imageFile := Image{
				AbsoluteURL: fullFilePath,
				Name:        img.Name(),
			}
			log.Printf("imageFile: %v", imageFile)
			imageSlice = append(imageSlice, &imageFile)
		} else {
			log.Printf("%s err: %v\n", fullFilePath, err)
		}
	}

	return imageSlice, nil
}

// fileIsImage returns true file the file contents is an image MIME-type.
// Otherwise returns false.
func fileIsImage(f *os.File) (bool, error) {
	buffer := make([]byte, 512)

	_, err := f.Read(buffer)
	if err != nil {
		return false, err
	}

	contentType := http.DetectContentType(buffer)
	isImage := strings.Contains(contentType, ImageContentType)
	return isImage, nil
}
