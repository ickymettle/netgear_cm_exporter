package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewConfigFromFile(t *testing.T) {
	want := &Config{
		Modem: Modem{
			Address:  "192.168.100.1",
			Username: "admin",
			Password: "foobaz",
		},
		Telemetry: Telemetry{
			ListenAddress: ":9527",
			MetricsPath:   "/metrics",
		},
	}

	got, err := NewConfigFromFile("testdata/minimal.yml")
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("config differs (-want, +got): %s", diff)
	}

}
