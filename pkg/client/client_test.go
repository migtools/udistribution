package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/google/go-cmp/cmp"
	uconfig "github.com/kaovilai/udistribution/pkg/distribution/configuration"
)

func TestNewClient(t *testing.T) {
	type args struct {
		configString string
		envs         []string
	}
	tests := []struct {
		name       string
		args       args
		wantClient *Client
		wantErr    bool
	}{
		{
			name: "empty config",
			args: args{
				configString: "",
				envs:         []string{},
			},
			wantClient: &Client{
				config: uconfig.GetWantConfig(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClient, err := NewClient(tt.args.configString, tt.args.envs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// TODO: compare app
			tt.wantClient.app = gotClient.app
			// if secret is empty, copy generated secret
			if tt.wantClient.config.HTTP.Secret == "" {
				tt.wantClient.config.HTTP.Secret = gotClient.config.HTTP.Secret
			}
			// TODO: remove ignore storage filesystem useragent
			tt.wantClient.config.Storage["filesystem"]["useragent"] = gotClient.config.Storage["filesystem"]["useragent"]
			// if not disable then enable per https://github.com/distribution/distribution/blob/f637481c67241151dc6d6fe2b12852e2ad8d70c2/registry/handlers/app.go#L225-L227
			if !tt.wantClient.config.Validation.Enabled {
				tt.wantClient.config.Validation.Enabled = !tt.wantClient.config.Validation.Disabled
			}
			if !reflect.DeepEqual(gotClient, tt.wantClient) {
				t.Errorf(cmp.Diff(gotClient.app, tt.wantClient.app))
				t.Errorf(cmp.Diff(gotClient.config, tt.wantClient.config))
				t.Errorf("NewClient() = %v, want %v", gotClient, tt.wantClient)
			}
		})
	}
}

func getContext(config *configuration.Configuration) context.Context {
	c, _ := GetContext(config)
	return c
}
