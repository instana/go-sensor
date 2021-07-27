// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk_test

import (
	"reflect"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
)

func TestLambdaClientContext_Base64(t *testing.T) {
	tests := []struct {
		name    string
		lc      instaawssdk.LambdaClientContext
		want    string
		wantErr bool
	}{
		{
			name: "Empty",
			lc:   instaawssdk.LambdaClientContext{},
			//{"Client":{"installation_id":"","app_title":"","app_version_code":"","app_package_name":""},"env":null,"custom":null}
			want:    "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiIiwiYXBwX3RpdGxlIjoiIiwiYXBwX3ZlcnNpb25fY29kZSI6IiIsImFwcF9wYWNrYWdlX25hbWUiOiIifSwiZW52IjpudWxsLCJjdXN0b20iOm51bGx9Cg==",
			wantErr: false,
		},
		{
			name: "Not empty",
			lc: instaawssdk.LambdaClientContext{
				Client: instaawssdk.LambdaClientContextClientApplication{
					InstallationID: "1234",
					AppTitle:       "MyApp",
					AppVersionCode: "1",
					AppPackageName: "My package",
				},
				Env: map[string]string{
					"env1": "1",
					"env2": "2",
				},
				Custom: map[string]string{
					instana.FieldT: "4",
					instana.FieldS: "5",
					instana.FieldL: "6",
				},
			},
			//{"Client":{"installation_id":"1234","app_title":"MyApp","app_version_code":"1","app_package_name":"My package"},"env":{"env1":"1","env2":"2"},"custom":{"x-instana-l":"6","x-instana-s":"5","x-instana-t":"4"}}
			want:    "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiMTIzNCIsImFwcF90aXRsZSI6Ik15QXBwIiwiYXBwX3ZlcnNpb25fY29kZSI6IjEiLCJhcHBfcGFja2FnZV9uYW1lIjoiTXkgcGFja2FnZSJ9LCJlbnYiOnsiZW52MSI6IjEiLCJlbnYyIjoiMiJ9LCJjdXN0b20iOnsieC1pbnN0YW5hLWwiOiI2IiwieC1pbnN0YW5hLXMiOiI1IiwieC1pbnN0YW5hLXQiOiI0In19Cg==",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.lc.Base64JSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64Json() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Base64Json() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLambdaClientContextFromBase64EncodedJSON(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    instaawssdk.LambdaClientContext
		wantErr bool
	}{
		{
			name:    "Empty",
			args:    "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiIiwiYXBwX3RpdGxlIjoiIiwiYXBwX3ZlcnNpb25fY29kZSI6IiIsImFwcF9wYWNrYWdlX25hbWUiOiIifSwiZW52IjpudWxsLCJjdXN0b20iOm51bGx9Cg==",
			want:    instaawssdk.LambdaClientContext{},
			wantErr: false,
		},
		{
			name: "Not Empty",
			args: "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiMTIzNCIsImFwcF90aXRsZSI6Ik15QXBwIiwiYXBwX3ZlcnNpb25fY29kZSI6IjEiLCJhcHBfcGFja2FnZV9uYW1lIjoiTXkgcGFja2FnZSJ9LCJlbnYiOnsiZW52MSI6IjEiLCJlbnYyIjoiMiJ9LCJjdXN0b20iOnsieC1pbnN0YW5hLWwiOiI2IiwieC1pbnN0YW5hLXMiOiI1IiwieC1pbnN0YW5hLXQiOiI0In19",
			want: instaawssdk.LambdaClientContext{
				Client: instaawssdk.LambdaClientContextClientApplication{
					InstallationID: "1234",
					AppTitle:       "MyApp",
					AppVersionCode: "1",
					AppPackageName: "My package",
				},
				Env: map[string]string{
					"env1": "1",
					"env2": "2",
				},
				Custom: map[string]string{
					instana.FieldT: "4",
					instana.FieldS: "5",
					instana.FieldL: "6",
				},
			},
			wantErr: false,
		},
		{
			name:    "Error",
			args:    "eyJDbAAAAAAAAAAAAAAiwiYXBwX3ZlcnNpb25fY29kZSI6IiIsImFwcF9wYAAAAAAAAAAAAAAAdXN0b20iOm51bGx9Cg==",
			want:    instaawssdk.LambdaClientContext{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instaawssdk.NewLambdaClientContextFromBase64EncodedJSON(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLambdaClientContextFromBase64EncodedJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLambdaClientContextFromBase64EncodedJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}
