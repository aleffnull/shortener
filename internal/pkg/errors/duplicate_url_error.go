package errors

import "fmt"

type DuplicateURLError struct {
	Key string
	URL string
}

func NewDuplicateURLError(key, url string) *DuplicateURLError {
	return &DuplicateURLError{
		Key: key,
		URL: url,
	}
}

func (e *DuplicateURLError) Error() string {
	return fmt.Sprintf("URL %v already exists with key %v", e.URL, e.Key)
}
