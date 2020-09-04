package filestore

import (
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Store defines a structure for watching a directory for image files changes,
// and getting the images in the directory.
type Store struct {
	list  FileList
	mutex sync.RWMutex
}

// Image represents an image on disk.
type Image struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// FileList is an aggregation of Images
type FileList map[string]Image

// New returns a new Store
func New() *Store {
	return &Store{
		list: make(FileList),
	}
}

// Watch starts the Store watching a specific directory.
func (f *Store) Watch(dir string) error {
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
				f.handleEvent(event)
				// TODO: decide on watcher.Errors handling
			}
		}
	}()

	return nil
}

func (f *Store) handleEvent(event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create:
		// TODO:implement add functionality.
	case fsnotify.Remove, fsnotify.Rename:
		//TODO: implement remove functionality.
	}
}

// Files returns the list of files current in the store.
func (f *Store) Files() FileList {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.list
}
