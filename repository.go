package shortly

import (
	"errors"
	"sync"
)

var (
	// ErrNoResults is used when there are no results in the repository
	// for the given ID.
	ErrNoResults = errors.New("no results found with that ID")
)

// Repository interface used to define the contract
// of a storage repository. Set should be used for
// storing a URLObject against an ID and Get should
// return a URLObject for a given ID. Exists allows for
// us to check if a URLObject already exists in the DB.
type Repository interface {
	Set(ID string, res *URLObject) error
	Get(ID string) (*URLObject, error)
	Exists(url string) (*URLObject, bool)
}

// mapRepository implements the Repository interface
// and uses an in-memory map to store results
// against a generated hash ID. To guard against concurrent
// reads and write we use a sync.RWMutex as an embedded type.
type mapRepository struct {
	sync.RWMutex
	data map[string]*URLObject
}

func newMapRepository() *mapRepository {
	return &mapRepository{
		data: make(map[string]*URLObject),
	}
}

// Set takes an ID and *URLObject, storing them in the map.
func (a *mapRepository) Set(ID string, res *URLObject) error {
	a.Lock()
	a.data[ID] = res
	a.Unlock()

	// since this is just assigning to a map, error is always nil.
	return nil
}

// Get takes a hash ID and returns either the associated
// *URLObject or an ErrNoResults.
func (a *mapRepository) Get(ID string) (*URLObject, error) {
	a.RLock()
	defer a.RUnlock()

	res, ok := a.data[ID]
	if !ok {
		return nil, ErrNoResults
	}

	return res, nil
}

// Exists loops through the map, checking if a result with
// a matching 'OriginalURL' exists already, meaning we don't
// need to generate a new one.
func (a *mapRepository) Exists(url string) (*URLObject, bool) {
	a.RLock()
	defer a.RUnlock()

	for _, v := range a.data {
		if v.OriginalURL == url {
			return v, true
		}
	}
	return nil, false
}
