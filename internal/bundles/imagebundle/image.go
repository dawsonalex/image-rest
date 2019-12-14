package imagebundle

import "net/url"

type (
	// Image represents meta-data associated with an image on the file system.
	Image struct {
		width       int
		height      int
		name        string
		absoluteURL *url.URL
	}
)
