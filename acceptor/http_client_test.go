// (c) Copyright IBM Corp. 2024

package acceptor_test

import (
	"os"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
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
			_, err := acceptor.NewHTTPClient(tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHTTPClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
