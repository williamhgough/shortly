package adding

import (
	"fmt"
	"log"
	"time"
)

// Service exposes the adding service interface
type Service interface {
	CreateShortURL(object *URLObject) (*URLObject, error)
}

// Repository interface used to define the contract
// of a storage repository. Set should be used for
// storing a URLObject against an ID and Get should
// return a URLObject for a given ID. Exists allows for
// us to check if a URLObject already exists in the DB.
type Repository interface {
	Set(ID string, res *URLObject) error
	Exists(url string) (*URLObject, bool)
}

// Hasher is a simple interface that should allow
// for us to swap out our chosen hashing library
// or method in the future.
type Hasher interface {
	Generate(salt string, t time.Time) (string, error)
}

// Service represents the actual shortly service
// and owns the hasher and data repository. Using
// the interfaces allows for the ability to run multiple
// instances of the service using different storage
// providers or hash generators.
type service struct {
	hasher Hasher
	db     Repository
}

func (s *service) CreateShortURL(object *URLObject) (*URLObject, error) {
	// Check the db to see if a shortURL already exists.
	// If so, return the associated short URL.
	// Uncomment this out if you aren't bothered about
	// one unique hash per URL.
	if res, exists := s.db.Exists(object.OriginalURL); exists {
		log.Printf("ID %s already exists for URL: %s", res.ID, res.OriginalURL)
		return res, nil
	}

	// Generate the short URL ID using the url and current time
	// as the hasher salt and unique value for encoding.
	hash, err := s.hasher.Generate(object.OriginalURL, time.Now())
	if err != nil {
		log.Printf("failed to generate short URL: %s\n", err)
		return nil, err
	}

	// Update the URLObject ID and ShortURL
	object.ID = hash
	object.ShortURL = fmt.Sprintf("%s://%s/%s", object.Scheme, object.Host, hash)

	// Add the URLObject to the database.
	err = s.db.Set(hash, object)
	if err != nil {
		log.Printf("failed to store result for ID %s: %s\n", hash, err)
		return nil, err
	}
	return object, nil
}

// NewService returns a new instance of an adding service
func NewService(r Repository, h Hasher) Service {
	return &service{
		hasher: h,
		db:     r,
	}
}
