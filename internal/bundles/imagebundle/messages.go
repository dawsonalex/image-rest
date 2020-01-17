package imagebundle

import (
	"github.com/dawsonalex/image-rest/internal/imagelibrary"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

/*
 * This file holds all the accepted JSON request and response types for the imagebundle as structs.
 */

// Describes a JSON response for a request for a single image.
type imageResponse struct {
	Name   string    `json:"name"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	ID     uuid.UUID `json:"id"`
}

// Describes a JSON response for a request for a group of images.
type imageLibraryResponse struct {
	Count   int             `json:"count"`
	Library []imageResponse `json:"library"`
}

// Describes an error response.
type errorResponse struct {
	Msg string `json:"msg"`
}

// ImageController contains a number of functions for handling requests
// regarding images or image meta-data.
type ImageController struct {
	ContentDir  string `json:"contentDir"` // The path to content. Defaults to user HOME
	images      map[uuid.UUID]*imagelibrary.Image
	libraryInit bool // true if the image library has been loaded, otherwise false.

	logger *log.Logger
}
