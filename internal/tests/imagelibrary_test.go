package tests

import (
	"os"
	"testing"

	"github.com/dawsonalex/image-rest/internal/imagelibrary"
)

func TestCanDetermineFileAreImages(t *testing.T) {
	filePath := "/home/ad/Pictures/wavy_triangles.png"
	if file, err := os.Open(filePath); err == nil {
		if isImage, err := imagelibrary.IsImageContentType(file); err != nil {
			t.Errorf("error: %v", err)
			t.FailNow()
		} else if isImage {
			t.Log("File is image")
		} else {
			t.Error("File isn't recognised as image")
			t.FailNow()
		}
	} else {
		t.Errorf("Cannot open file %s: %v", filePath, err)
		t.FailNow()
	}
}
