package configuration

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/google/go-cmp/cmp"
	def "github.com/kaovilai/udistribution/pkg/client/default"
)

func TestParseEnvironment(t *testing.T) {
	type args struct {
		configString string
		envs         []string
	}
	tests := []struct {
		name       string
		args       args
		wantConfig *configuration.Configuration
		wantErr    bool
	}{
		{
			name: "empty config",
			args: args{
				configString: "",
				envs:         []string{},
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "default config",
			args: args{
				configString: def.Config,
				envs:         []string{},
			},
			wantConfig: &defaultWantConfig,
		},
		{
			name: "default config with s3 env",
			args: args{
				configString: def.Config,
				envs: []string{
					"REGISTRY_STORAGE=s3",
					"REGISTRY_STORAGE_S3_BUCKET=test-bucket",
					"REGISTRY_STORAGE_S3_REGION=us-east-1",
					"REGISTRY_STORAGE_S3_ACCESSKEYID=AKIAIOSFODNN7EXAMPLE",
					"REGISTRY_STORAGE_S3_SECRETACCESSKEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					"REGISTRY_STORAGE_S3_ENDPOINT=http://test-endpoint:4572",
					"REGISTRY_STORAGE_S3_USE_HTTP=true",
					"REGISTRY_HTTP_ADDR=localhost:6000",
				},
			},
			wantConfig: GetWantConfig(
				WithStorage(configuration.Storage{
					"s3": configuration.Parameters{
						"accesskeyid":     string("AKIAIOSFODNN7EXAMPLE"),
						"bucket":          string("test-bucket"),
						"endpoint":        string("http://test-endpoint:4572"),
						"region":          string("us-east-1"),
						"secretaccesskey": string("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
						"use":             map[string]interface{}{"http": bool(true)},
					},
				}),
				WithHTTP(HTTP{Addr: "localhost:6000"}),
				WithHeaders(http.Header{"X-Content-Type-Options": []string{"nosniff"}}),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := ParseEnvironment(tt.args.configString, tt.args.envs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotConfig, tt.wantConfig) {
				t.Errorf(cmp.Diff(gotConfig, tt.wantConfig))
				t.Errorf("ParseEnvironment() = %v, want %v", gotConfig, tt.wantConfig)
			}
		})
	}
}
