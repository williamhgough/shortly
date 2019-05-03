package rest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/williamhgough/shortly/pkg/adding"
	"github.com/williamhgough/shortly/pkg/redirect"
	"github.com/williamhgough/shortly/pkg/storage/memory"
)

// Handler is responsible for routing requests to the appropriate service
func Handler(a adding.Service, r redirect.Service) http.Handler {
	router := httprouter.New()
	router.POST("/api/v1/shorten", generateURLHandler(a))
	router.GET("/:id", redirectHandler(r))
	return router
}

// generateURLHandler will accept a simple post body containing the original
// URL to be shortened, it will then use the configured hasher to generate a
// short URL, store the result in the repository and then return a
// ShortenResponse to the consumer.
func generateURLHandler(a adding.Service) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		var req adding.URLObject
		if err = json.Unmarshal(body, &req); err != nil {
			log.Printf("failed to unmarshal request body: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		req.Host = r.URL.Host
		req.Scheme = r.URL.Scheme

		res, err := a.CreateShortURL(&req)
		if err != nil {
			log.Printf("failed to create short URL: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respond(w, res)
	}
}

// redirectHandler is responsible for accepting a shortened URL.
// If there is an associated ID the consumer is redirected.
// If there is no associated ID the service returns http.StatusNotFound.
func redirectHandler(redirecter redirect.Service) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Printf("[%s] %s\n", r.Method, r.URL.Path)

		// Check that the request method is correct
		if r.Method != http.MethodGet {
			log.Printf("incorrect request method type: %s, expecting GET\n", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Fetch ID from querystring paramaters
		ID := p.ByName("id")
		if ID == "" {
			log.Printf("No hash path given, can't redirect\n")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Once we have the hash we can fetch the associated
		// result and continue on to redirect.
		results, err := redirecter.GetShortURL(ID)
		if err == memory.ErrNoResults {
			log.Printf("no short URL for the given ID: %s\n", ID)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Redirect the consumer to their original URL.
		http.Redirect(w, r, results.OriginalURL, http.StatusMovedPermanently)
	}
}

// respond is used to marshal and write out ShortenResponse data.
func respond(w http.ResponseWriter, data *adding.URLObject) {
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
