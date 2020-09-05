package imageservice

import (
	"image"
	_ "image/gif"  // Register gif image decoding
	_ "image/jpeg" // Register jpeg image decoding
	_ "image/png"  // Register PNG image decoding
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

const imageContentType string = "image/"

// Service defines a structure for watching a directory for image files changes,
// and getting the images in the directory.
type Service struct {
	log   *logrus.Logger
	list  ImageList
	mutex sync.RWMutex
}

// Image represents an image on disk.
type Image struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ImageList is an aggregation of Images
type ImageList map[string]Image

// New returns a reference to a new Service, or a nil reference and
// and an error if something goes wrong setting up the store for
// the given directory.
func New(logger *logrus.Logger) *Service {
	return &Service{
		log: logger,
	}
}

func loadFiles(dir string) (ImageList, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	images := make(ImageList, 0)
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		image, err := loadImage(path)
		if err != nil {
			continue
		}
		images[image.Name] = *image
	}
	return images, nil
}

func loadImage(filename string) (*Image, error) {
	reader, err := os.Open(filename)
	defer reader.Close()
	if err != nil {
		return nil, err
	}

	imageConfig, _, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, err
	}

	return &Image{
		Name:   filepath.Base(reader.Name()),
		Width:  imageConfig.Width,
		Height: imageConfig.Height,
	}, nil
}

// Watch starts the Service watching a specific directory.
func (s *Service) Watch(dir string) error {
	// load all images from the directory
	images, err := loadFiles(dir)
	if err != nil {
		return err
	}
	s.list = images

	// begin watching the directory for changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err = watcher.Add(dir); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				s.handleEvent(event)
				// TODO: decide on watcher.Errors handling
			}
		}
	}()

	return nil
}

func (s *Service) add(filename string) {
	image, err := loadImage(filename)
	if err != nil {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.list[image.Name] = *image
}

func (s *Service) remove(filename string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.list, filepath.Base(filename))
}

func (s *Service) handleEvent(event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create:
		s.add(event.Name)
	case fsnotify.Remove, fsnotify.Rename:
		//TODO: implement remove functionality.
	}
}

// Files returns the list of files current in the store.
func (s *Service) Files() ImageList {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.list
}
