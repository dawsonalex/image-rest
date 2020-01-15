package imagebundle

import (
	"net/url"

	"github.com/google/uuid"
)

type (
	// Image represents meta-data associated with an image on the file system.
	Image struct {
		Width       int       `json:"width"`
		Height      int       `json:"height"`
		Name        string    `json:"name"`
		AbsoluteURL *url.URL  `json:"url"`
		ID          uuid.UUID `json:"id"`
	}
)
