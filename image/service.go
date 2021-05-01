package image

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	_ "image/gif"  // Register gif image decoding
	_ "image/jpeg" // Register jpeg image decoding
	_ "image/png"  // Register PNG image decoding
)

// Service defines a structure for watching a directory for image files changes,
// and getting the images in the directory.
type Service struct {
	log     *logrus.Logger
	store   *Store
	watcher *fsnotify.Watcher
	stop    chan struct{}
}

// New returns a reference to a new Service, or a nil reference and
// and an error if something goes wrong setting up the store for
// the given directory.
func New(logger *logrus.Logger) *Service {
	return &Service{
		log: logger,
	}
}

// Watch starts the Service watching a specific directory.
func (s *Service) Watch(dir string) error {
	// load all images from the directory
	images, err := fromDir(dir)
	if err != nil {
		return err
	}
	s.store.AddAll(images)

	// begin watching the directory for changes
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = w

	if err = s.watcher.Add(dir); err != nil {
		return err
	}
	go func() {
		for {
			event, ok := <-s.watcher.Events
			if !ok {
				return
			}
			s.handleEvent(event)
		}
	}()

	go func() {
		for {
			err, ok := <-s.watcher.Errors
			if !ok {
				return
			}
			if err != nil {
				s.log.Errorf("image.Watch(): %v", err)
			}
		}
	}()

	return nil
}

// Stop makes the service finish watching the directory, and
// cleanup resources.
func (s *Service) Stop() {
	s.log.Println("image.Stop(): stopping image service")
	s.watcher.Close()
}

func (s *Service) add(filename string) {
	image, err := fromFile(filename)
	if err != nil {
		s.log.Errorf("image.add(): error loading image %s: %v", filename, err)
		return
	}
	s.log.Debugf("image.add(): adding image %s", filename)
	s.store.Add(*image)
}

func (s *Service) handleEvent(event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create:
		s.log.Printf("image.handleEvent(): handling event %v on %s", event.Op, event.Name)
		s.add(event.Name)
	case fsnotify.Remove, fsnotify.Rename:
		s.log.Printf("image.handleEvent(): handling event %v on %s", event.Op, event.Name)
		s.store.Remove(event.Name)
	}
}

// Files returns the list of files current in the store.
func (s *Service) Files() []*Image {
	return s.store.List()
}
