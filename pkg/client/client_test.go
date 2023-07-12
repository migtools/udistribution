package client

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/google/go-cmp/cmp"
	uconfig "github.com/migtools/udistribution/pkg/distribution/configuration"
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
		{
			name: "empty config with s3 env",
			args: args{
				configString: "",
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
			wantClient: &Client{
				config: uconfig.GetWantConfig(
					uconfig.WithStorage(configuration.Storage{
						"s3": configuration.Parameters{
							"accesskeyid":     string("AKIAIOSFODNN7EXAMPLE"),
							"bucket":          string("test-bucket"),
							"endpoint":        string("http://test-endpoint:4572"),
							"region":          string("us-east-1"),
							"secretaccesskey": string("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
							"use":             map[string]interface{}{"http": bool(true)},
						},
					}),
					uconfig.WithHTTP(uconfig.HTTP{Addr: "localhost:6000"}),
					uconfig.WithHeaders(http.Header{"X-Content-Type-Options": []string{"nosniff"}}),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotClient, err := NewClient(tt.args.configString, tt.args.envs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			rr := httptest.NewRecorder()
			rq, err := http.NewRequest("GET", "/v2/", strings.NewReader(""))
			if err != nil {
				t.Fatal(err)
			}
			gotClient.GetApp().ServeHTTP(rr, rq)
			if rr.Result().StatusCode != http.StatusOK && !tt.wantErr {
				t.Errorf("NewClient() = %v, want %v", rr.Result().StatusCode, http.StatusOK)
			}
			// TODO: compare app
			tt.wantClient.app = gotClient.app
			// if secret is empty, copy generated secret
			if tt.wantClient.config.HTTP.Secret == "" {
				tt.wantClient.config.HTTP.Secret = gotClient.config.HTTP.Secret
			}
			// TODO: remove ignore storage filesystem useragent
			for i, _ := range gotClient.config.Storage {
				if gotClient.config.Storage[i]["useragent"] != nil && gotClient.config.Storage[i]["useragent"] != "" {
					tt.wantClient.config.Storage[i]["useragent"] = gotClient.config.Storage[i]["useragent"]
				}
			}
			// if not disable then enable per https://github.com/distribution/distribution/blob/f637481c67241151dc6d6fe2b12852e2ad8d70c2/registry/handlers/app.go#L225-L227
			if !tt.wantClient.config.Validation.Enabled {
				tt.wantClient.config.Validation.Enabled = !tt.wantClient.config.Validation.Disabled
			}
			if !reflect.DeepEqual(gotClient, tt.wantClient) {
				// t.Errorf(cmp.Diff(gotClient.app, tt.wantClient.app))
				t.Errorf(cmp.Diff(gotClient.config, tt.wantClient.config))
				t.Errorf("NewClient() = %v, want %v", gotClient, tt.wantClient)
			}
		})
	}
}

// TODO: fix test
// func TestHTTPResponses(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req *http.Request
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *http.Response
// 		wantErr bool
// 	}{
// 		{
// 			name: "test http responses",
// 			args: args{
// 				ctx: context.Background(),
// 				req: &http.Request{
// 					Method: "GET",
// 					URL:    &url.URL{
// 						Path: "/v2/",
// 					},
// 				},
// 			},
// 			want: &http.Response{
// 				StatusCode: 200,
// 				Header: http.Header{
// 					"Content-Type": []string{"application/json; charset=utf-8"},
// 					"Docker-Experimental": []string{"true"},
// 					"X-Docker-Registry-Version": []string{"0.0.0"},
// 					"X-Docker-Token": []string{"true"},
// 					"X-Docker-Endpoints": []string{"http://localhost:5000"},
// 					"X-Docker-Location": []string{"http://localhost:5000/v2/"},
// 					"X-Docker-Mirror": []string{"https://registry-1.docker.io"},
// 					"X-Docker-V2-Support": []string{"true"},
// 					"X-Docker-Endpoints-Path": []string{"/v2/"},
// 					"X-Docker-Token-Expires": []string{"Thu, 01 Jan 1970 00:00:00 GMT"},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			rr := httptest.NewRecorder()
// 			// init client
// 			client, err := NewClient("", []string{})
// 			if err != nil {
// 				t.Errorf("NewClient() error = %v", err)
// 				return
// 			}
// 			rw := http.ResponseWriter(rr)
// 			client.GetApp().ServeHTTP(rr, tt.args.req)
// 			if (rr.Result().StatusCode != tt.want.StatusCode) != tt.wantErr {
// 				t.Errorf("GetApp().ServeHTTP() = %v, want %v", rr.Result().StatusCode, tt.want.StatusCode)
// 			}
// 			got, err := tt.args.req.Do(tt.args.ctx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Request.Do() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Request.Do() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
