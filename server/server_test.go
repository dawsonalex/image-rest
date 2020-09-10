package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

const sampleDir = "../sample_images"

func TestUploadHander(t *testing.T) {
	files, err := ioutil.ReadDir(sampleDir)
	if err != nil {
		t.Skipf("could not read directory to compare results: %v", err)
	}
	fileCountBeforeRequest := len(files)

	request, _ := http.NewRequest("GET", "/files", nil)
	response := httptest.NewRecorder()
	UploadHandler(sampleDir, logrus.New())(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("error, got response code: %d", response.Code)
	}

	files, err = ioutil.ReadDir(sampleDir)
	if err != nil {
		t.Skipf("could not read directory to compare results: %v", err)
	}

}

func TestFilesHandler(t *testing.T) {
	request, _ := http.NewRequest("GET", "/files", nil)
	response := httptest.NewRecorder()

	FilesHandler()

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
