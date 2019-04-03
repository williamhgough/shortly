package shortly

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var s *Service

func init() {
	s = New(8080)
	go s.Start()
}

func TestService_generateURLHandler(t *testing.T) {
	// set some dummy data
	s.db.Set("123456", &URLObject{
		ID:          "123456",
		OriginalURL: "http://google.com",
		ShortURL:    "http://short.ly/123456",
	})

	tests := []struct {
		name           string
		method         string
		input          io.Reader
		expectedStatus int
	}{
		{
			"can generate a short URL",
			http.MethodPost,
			strings.NewReader(`{"original_url":"http://google.co.uk"}`),
			http.StatusOK,
		},
		{
			"returns existing short URL",
			http.MethodPost,
			strings.NewReader(`{"original_url":"http://google.com"}`),
			http.StatusOK,
		},
		{
			"Incorrect method not allowed",
			http.MethodGet,
			strings.NewReader(`{"original_url":"http://google.com"}`),
			http.StatusMethodNotAllowed,
		},
		{
			"unable to unmarshal returns 500",
			http.MethodPost,
			strings.NewReader(`{"text":}`),
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/api/v1/shorten", tt.input)
			r.URL.Scheme = "http"
			r.URL.Host = "short.ly"
			s.generateURLHandler(w, r)

			if w.Result().StatusCode != tt.expectedStatus {
				t.Logf("got %d, wanted: %d", w.Result().StatusCode, tt.expectedStatus)
				t.Fail()
			}
		})
	}
}

func TestService_redirectHandler(t *testing.T) {
	// set some dummy data
	s.db.Set("123456", &URLObject{
		ID:          "123456",
		OriginalURL: "http://google.com",
		ShortURL:    "http://short.ly/123456",
	})

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			"can redirect to original url",
			http.MethodGet,
			"/123456",
			http.StatusMovedPermanently,
		},
		{
			"Incorrect method not allowed",
			http.MethodPost,
			"/123456",
			http.StatusMethodNotAllowed,
		},
		{
			"No ID given returns http.StatusBadRequest",
			http.MethodGet,
			"/",
			http.StatusBadRequest,
		},
		{
			"Non-existent ID gives http.StatusNoContent",
			http.MethodGet,
			"/abcdef",
			http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.url, nil)
			s.redirectHandler(w, r)

			if w.Result().StatusCode != tt.expectedStatus {
				t.Logf("got %d, wanted: %d", w.Result().StatusCode, tt.expectedStatus)
				t.Fail()
			}
		})
	}
}
