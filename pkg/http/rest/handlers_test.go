package rest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/williamhgough/shortly/pkg/adding"
	"github.com/williamhgough/shortly/pkg/hashing"
	"github.com/williamhgough/shortly/pkg/redirect"
	"github.com/williamhgough/shortly/pkg/storage/memory"
)

var (
	mem        = memory.NewMapRepository()
	adder      = adding.NewService(mem, hashing.NewSimpleHasher())
	redirecter = redirect.NewService(mem)
)

func TestService_generateURLHandler(t *testing.T) {
	mem.Set("123456", &adding.URLObject{
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
			p := httprouter.Params{}

			generateURLHandler(adder)(w, r, p)

			t.Logf(w.Body.String())

			if w.Result().StatusCode != tt.expectedStatus {
				t.Logf("got %d, wanted: %d", w.Result().StatusCode, tt.expectedStatus)
				t.Fail()
			}
		})
	}
}

func TestService_redirectHandler(t *testing.T) {
	mem.Set("123456", &adding.URLObject{
		ID:          "123456",
		OriginalURL: "http://google.com",
		ShortURL:    "http://short.ly/123456",
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		ID             string
	}{
		{
			"can redirect to original url",
			http.MethodGet,
			http.StatusMovedPermanently,
			"123456",
		},
		{
			"Incorrect method not allowed",
			http.MethodPost,
			http.StatusMethodNotAllowed,
			"123456",
		},
		{
			"No ID given returns http.StatusBadRequest",
			http.MethodGet,
			http.StatusBadRequest,
			"",
		},
		{
			"Non-existent ID gives http.StatusNoContent",
			http.MethodGet,
			http.StatusNoContent,
			"abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/"+tt.ID, nil)

			redirectHandler(redirecter)(w, r, httprouter.Params{{Key: "id", Value: tt.ID}})

			if w.Result().StatusCode != tt.expectedStatus {
				t.Logf("got %d, wanted: %d", w.Result().StatusCode, tt.expectedStatus)
				t.Fail()
			}
		})
	}
}
