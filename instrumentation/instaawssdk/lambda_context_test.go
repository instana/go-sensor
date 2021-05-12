package instaawssdk

import "testing"

func TestLambdaClientContext_Base64(t *testing.T) {
	tests := []struct {
		name    string
		lc      LambdaClientContext
		want    string
		wantErr bool
	}{
		{
			name: "Empty",
			lc:   LambdaClientContext{},
			//{"Client":{"installation_id":"","app_title":"","app_version_code":"","app_package_name":""},"env":null,"custom":null}
			want:    "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiIiwiYXBwX3RpdGxlIjoiIiwiYXBwX3ZlcnNpb25fY29kZSI6IiIsImFwcF9wYWNrYWdlX25hbWUiOiIifSwiZW52IjpudWxsLCJjdXN0b20iOm51bGx9Cg==",
			wantErr: false,
		},
		{
			name: "Empty",
			lc: LambdaClientContext{
				Client: LambdaClientContextClientApplication{
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
					"X-INSTANA-T": "4",
					"X-INSTANA-S": "5",
					"X-INSTANA-L": "6",
				},
			},
			//{"Client":{"installation_id":"1234","app_title":"MyApp","app_version_code":"1","app_package_name":"My package"},"env":{"env1":"1","env2":"2"},"custom":{"X-INSTANA-L":"6","X-INSTANA-S":"5","X-INSTANA-T":"4"}}
			want:    "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiMTIzNCIsImFwcF90aXRsZSI6Ik15QXBwIiwiYXBwX3ZlcnNpb25fY29kZSI6IjEiLCJhcHBfcGFja2FnZV9uYW1lIjoiTXkgcGFja2FnZSJ9LCJlbnYiOnsiZW52MSI6IjEiLCJlbnYyIjoiMiJ9LCJjdXN0b20iOnsiWC1JTlNUQU5BLUwiOiI2IiwiWC1JTlNUQU5BLVMiOiI1IiwiWC1JTlNUQU5BLVQiOiI0In19Cg==",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.lc.Base64Json()
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
