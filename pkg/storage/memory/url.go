package memory

// URLObject represents the saved information in the DB.
// It contains the generated hash ID, the original URL
// and the newly generated short URL.
type URLObject struct {
	ID          string `json:"id,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	ShortURL    string `json:"short_url,omitempty"`
}
