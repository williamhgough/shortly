package shortly

import (
	"testing"
	"time"
)

func Test_simpleHasher_New(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2019-02-24T11:24:01+00:00")
	if err != nil {
		t.Logf("could not generate time for testing")
		t.Fail()
	}

	testTime2, err := time.Parse(time.RFC3339, "2019-02-24T12:23:01+00:00")
	if err != nil {
		t.Logf("could not generate time for testing")
		t.Fail()
	}

	hasher := newSimpleHasher()

	type args struct {
		salt string
		t    time.Time
	}
	tests := []struct {
		name    string
		s       *simpleHasher
		args    args
		want    string
		wantErr bool
	}{
		{"generates hash as expected", hasher, args{"http://google.com", testTime}, "gLx3oLG", false},
		{"same data generates same hash", hasher, args{"http://google.com", testTime}, "gLx3oLG", false},
		{"same time different salt creates different hash", hasher, args{"http://google.co.uk", testTime}, "gOgRoOx", false},
		{"different time same salt creates different hash", hasher, args{"http://google.co.uk", testTime2}, "1obD0a9", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &simpleHasher{}
			got, err := s.Generate(tt.args.salt, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("simpleHasher.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("simpleHasher.New() = %v, want %v", got, tt.want)
			}
		})
	}
}
