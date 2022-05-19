package client

import (
	"reflect"
	"testing"

	"github.com/distribution/distribution/v3/configuration"
)

func Test_getDefaultConfig(t *testing.T) {
	tests := []struct {
		name       string
		wantConfig *configuration.Configuration
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotConfig := getDefaultConfig(); !reflect.DeepEqual(gotConfig, tt.wantConfig) {
				t.Errorf("getDefaultConfig() = %v, want %v", gotConfig, tt.wantConfig)
			}
		})
	}
}


