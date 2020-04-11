package imagelibrary

const errFileNotImageMsg = "file is not an image"

// ErrFileNotImage defines a type of error
type ErrFileNotImage struct {
	filePath string
}

// Error returns the message for this error.
func (ErrFileNotImage) Error() string {
	return errFileNotImageMsg
}

// FilePath returns the path to the file that caused the error.
func (err ErrFileNotImage) FilePath() string {
	return err.filePath
}
