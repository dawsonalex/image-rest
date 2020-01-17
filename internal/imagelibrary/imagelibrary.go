package imagelibrary

import (
	"image"
	_ "image/gif"  // Register gif image decoding
	_ "image/jpeg" // Register jpeg image decoding
	_ "image/png"  // Register PNG image decoding
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
)

// FromDir returns a slice of images at the path dir.
func FromDir(dir string) ([]*Image, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return make([]*Image, 0), err
	}

	imageSlice := make([]*Image, 0)
	log.Println(len(files))
	for _, img := range files {
		fullFilePath := filepath.Join(dir, img.Name())
		if reader, err := os.Open(fullFilePath); err == nil {
			defer reader.Close()
			imgConfig, _, err := image.DecodeConfig(reader)
			if err != nil {
				log.Printf("%s err: %v\n", fullFilePath, err)
				continue
			}

			imageFile := Image{
				AbsoluteURL: fullFilePath,
				Name:        img.Name(),
				Width:       imgConfig.Width,
				Height:      imgConfig.Height,
			}
			imageSlice = append(imageSlice, &imageFile)
		} else {
			log.Printf("%s err: %v\n", fullFilePath, err)
		}
	}

	return imageSlice, nil
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
	isImage := strings.Contains(contentType, ImageContentType)
	return isImage, nil
}
