package image

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

const sampleDir = "../sample_images"

func TestWatchDir(t *testing.T) {
	store := New(logrus.New())
	err := store.Watch(sampleDir)
	if err != nil {
		t.Error(err)
	}

	files, err := ioutil.ReadDir(sampleDir)
	if err != nil {
		t.Skipf("could not read directory to compare results: %v", err)
	}

	actualFileCount := len(files)
	storeFileCount := len(store.Files())
	if actualFileCount != storeFileCount {
		t.Errorf("expected store to contain %d files, but only found %d", actualFileCount, storeFileCount)
	}
}
