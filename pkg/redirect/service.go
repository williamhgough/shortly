package redirect

// Service exposes the redirect service interface
type Service interface {
	GetShortURL(ID string) (*URLObject, error)
}

// Repository interface used to define the contract
// of a storage repository. Set should be used for
// storing a URLObject against an ID and Get should
// return a URLObject for a given ID. Exists allows for
// us to check if a URLObject already exists in the DB.
type Repository interface {
	Get(ID string) (*URLObject, error)
}

// service represents the actual shortly service
// and owns the hasher and data repository. Using
// the interfaces allows for the ability to run multiple
// instances of the service using different storage
// providers or hash generators.
type service struct {
	db Repository
}

func (s *service) GetShortURL(ID string) (*URLObject, error) {
	res, err := s.db.Get(ID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// NewService returns a new instance of an adding service
func NewService(r Repository) Service {
	return &service{r}
}
