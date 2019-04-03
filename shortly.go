package shortly

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Service represents the actual shortly service
// and owns the hasher and data repository. Using
// the interfaces allows for the ability to run multiple
// instances of the service using different storage
// providers or hash generators.
type Service struct {
	port   uint
	hasher Hasher
	db     Repository
}

// URLObject represents the saved information in the DB.
// It contains the generated hash ID, the original URL
// and the newly generated short URL.
type URLObject struct {
	ID          string `json:"id,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	ShortURL    string `json:"short_url,omitempty"`
}

// New exposes the entrypoint of the services
// to any consumers. It returns a service instance
// with the configured hasher, repository and provided
// port to run the server on.
func New(port uint) *Service {
	return &Service{
		port:   port,
		hasher: newSimpleHasher(),
		db:     newMapRepository(),
	}
}

// Start bootstraps the service by adding the handlers
// to the default http.ServeMux and starting the server
// on the service instances' configured port. Will return
// an error if it fails to start.
func (s *Service) Start() error {
	http.HandleFunc("/", s.redirectHandler)
	http.HandleFunc("/api/v1/shorten", s.generateURLHandler)

	log.Printf("running short.ly server on http://localhost:%d", s.port)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", s.port), nil); err != nil {
		return fmt.Errorf("error starting short.ly server: %s", err)
	}

	return nil
}

// redirectHandler is responsible for accepting a shortened URL.
// If there is an associated ID the consumer is redirected.
// If there is no associated ID the service returns http.StatusNotFound.
func (s *Service) redirectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s\n", r.Method, r.URL.Path)

	// Check that the request method is correct
	if r.Method != http.MethodGet {
		log.Printf("incorrect request method type: %s, expecting GET\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Fetch ID from querystring paramaters
	ID := strings.TrimPrefix(r.URL.Path, "/")
	if ID == "" {
		log.Printf("No hash path given, can't redirect\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Once we have the hash we can fetch the associated
	// result and continue on to redirect.
	results, err := s.db.Get(ID)
	if err == ErrNoResults {
		log.Printf("no short URL for the given ID: %s\n", ID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Redirect the consumer to their original URL.
	http.Redirect(w, r, results.OriginalURL, http.StatusMovedPermanently)
}

// generateURLHandler will accept a simple post body containing the original
// URL to be shortened, it will then use the configured hasher to generate a
// short URL, store the result in the repository and then return a
// ShortenResponse to the consumer.
func (s *Service) generateURLHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s\n", r.Method, r.URL.Path)

	// Check that the request method is correct.
	if r.Method != http.MethodPost {
		log.Printf("incorrect request method type: %s, expecting POST\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body, if failed return 400.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("could not read request body: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Unmarshal the request body
	var req URLObject
	if err = json.Unmarshal(body, &req); err != nil {
		log.Printf("failed to unmarshal request body: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check the db to see if a shortURL already exists.
	// If so, return the associated short URL.
	// Uncomment this out if you aren't bothered about
	// one unique hash per URL.
	if res, exists := s.db.Exists(req.OriginalURL); exists {
		log.Printf("ID %s already exists for URL: %s", res.ID, res.OriginalURL)
		respond(w, res)
		return
	}

	// Generate the short URL ID using the url and current time
	// as the hasher salt and unique value for encoding.
	hash, err := s.hasher.Generate(req.OriginalURL, time.Now())
	if err != nil {
		log.Printf("failed to generate short URL: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update the URLObject ID and ShortURL
	req.ID = hash
	req.ShortURL = fmt.Sprintf("http://localhost:%d/%s", s.port, hash)

	// Add the URLObject to the database.
	err = s.db.Set(hash, &req)
	if err != nil {
		log.Printf("failed to store result for ID %s: %s\n", hash, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the URLObject as JSON.
	respond(w, &req)
}

// respond is used to marshal and write out ShortenResponse data.
func respond(w http.ResponseWriter, data *URLObject) {
	// Marshal the response body ready to write it back to the consumer
	response, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal response body: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the marshalled response and set appropriate headers
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
