package image

import (
	"path/filepath"
	"sync"
)

type Store struct {
	list  List
	mutex sync.RWMutex
}

// AddAll adds a list of images to the store.
func (s *Store) AddAll(images List) {
	for _, image := range images {
		s.Add(*image)
	}
}

// Add adds an image to the store.
func (s *Store) Add(image Image) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.list[image.Name] = &image
}

// Remove removes an images with the corresponding name from the store.
func (s *Store) Remove(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.list, filepath.Base(name))
}

// List returns a slice of all images in the store.
func (s *Store) List() []*Image {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	images := make([]*Image, 0)
	for _, image := range s.list {
		images = append(images, image)
	}
	return images
}
