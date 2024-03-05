// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor

import (
	"os"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	type args struct {
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				timeout: 500 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				timeout: 500 * time.Millisecond,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if !tt.wantErr {
			os.Setenv("INSTANA_ENDPOINT_PROXY", "localhost")
		} else {
			os.Setenv("INSTANA_ENDPOINT_PROXY", "http://{example.com")
		}
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHTTPClient(tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHTTPClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
