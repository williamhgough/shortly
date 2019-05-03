package hashing

import (
	"fmt"
	"time"

	"github.com/speps/go-hashids"
)

// Service is a simple interface that should allow
// for us to swap out our chosen hashing library
// or method in the future.
type Service interface {
	Generate(salt string, t time.Time) (string, error)
}

// Simple hasher implements the Service interface and
// wraps the following library:
//  - https://github.com/speps/go-hashids
type simpleHasher struct{}

// NewSimpleHasher returns a pointer to a new instance
// of a simpleHasher. Method exists to support reducing
// code changes if simpleHasher changes under the hood.
// For example, if a new dependency was added to the struct.
func NewSimpleHasher() Service {
	return &simpleHasher{}
}

// New takes an original URL and uses it as the salt
// for the hash generator, the current unix timestamp
// is then passed to the hash encoder, this should
// guarantee a unique short URL for each string given.
func (s *simpleHasher) Generate(salt string, t time.Time) (string, error) {
	// Set up new hasher
	hd := hashids.NewData()
	// Set the salt to the one provided
	hd.Salt = salt

	// create new HashID
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", fmt.Errorf("failed to create new hash: %s", err)
	}

	// Encode the given time as Unix timestamp to create the ID.
	ID, err := h.Encode([]int{int(t.Unix())})
	if err != nil {
		return "", fmt.Errorf("failed to encode hash: %s", err)
	}

	return ID, nil
}
