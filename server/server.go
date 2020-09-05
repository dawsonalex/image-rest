package server

import (
	"encoding/json"
	"net/http"

	"github.com/dawsonalex/image-rest/imageservice"
	"github.com/sirupsen/logrus"
)

// FilesHandler returns a http.HandlerFunc that accepts requests for an image stores
// file list.
func FilesHandler(store *imageservice.Service, logger *logrus.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.Errorf("Invalid HTTP method, got: %v", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		images := store.Files()
		imageResponse := make([]imageservice.Image, 0)
		for _, image := range images {
			imageResponse = append(imageResponse, image)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(imageResponse)
	})
}
