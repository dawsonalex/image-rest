package image

import (
	"image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Image represents an image on disk.
type Image struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// List is an aggregation of Images
type List map[string]*Image

// fromDir reads the directory `dir` and returns a list
// images it contains.
func fromDir(dir string) (List, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	images := make(List, 0)
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		img, err := fromFile(path)
		if err != nil {
			continue
		}
		images[img.Name] = img
	}
	return images, nil
}

func fromFile(filename string) (*Image, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if err != nil {
		return nil, err
	}

	return FromReader(reader, reader.Name())
}

func FromReader(reader io.Reader, name string) (*Image, error) {
	imageConfig, _, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, err
	}

	return &Image{
		Name:   filepath.Base(name),
		Width:  imageConfig.Width,
		Height: imageConfig.Height,
	}, nil
}
