package memory

import (
	"reflect"
	"testing"

	"github.com/williamhgough/shortly/pkg/redirect"

	"github.com/williamhgough/shortly/pkg/adding"
)

func Test_mapRepository_Set(t *testing.T) {
	a := NewMapRepository()
	a.Set("12345", &adding.URLObject{
		ID:          "12345",
		OriginalURL: "http://google.com",
		ShortURL:    "http://short.ly/12345",
	})

	if _, err := a.Get("12345"); err != nil {
		t.Fail()
	}
}

func Test_mapRepository_Get(t *testing.T) {
	a := NewMapRepository()
	res := &adding.URLObject{
		ID:          "12345",
		OriginalURL: "http://google.com",
		ShortURL:    "http://short.ly/12345",
	}
	a.Set("12345", res)

	type args struct {
		ID string
	}
	tests := []struct {
		name    string
		args    args
		want    *redirect.URLObject
		wantErr bool
	}{
		{"Can fetch stored results", args{ID: "12345"}, &redirect.URLObject{ID: "12345", OriginalURL: "http://google.com"}, false},
		{"Returns appropriate error if no results", args{ID: "123456"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.Get(tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapRepository.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapRepository.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapRepository_Exists(t *testing.T) {
	a := NewMapRepository()
	res := &adding.URLObject{
		ID:          "12345",
		OriginalURL: "http://google.com",
		ShortURL:    "http://short.ly/12345",
	}
	a.Set("12345", res)

	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"can find existing results", args{url: "http://google.com"}, true},
		{"returns false if no entry", args{url: "http://google.co.uk"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, exists := a.Exists(tt.args.url); exists != tt.want {
				t.Errorf("mapRepository.Exists() = %v, want %v", exists, tt.want)
			}
		})
	}
}
