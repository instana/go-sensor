package instaawssdk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

type LambdaClientContextClientApplication struct {
	InstallationID string `json:"installation_id"`
	AppTitle       string `json:"app_title"`
	AppVersionCode string `json:"app_version_code"`
	AppPackageName string `json:"app_package_name"`
}

type LambdaClientContext struct {
	Client LambdaClientContextClientApplication
	Env    map[string]string `json:"env"`
	Custom map[string]string `json:"custom"`
}

func (lc *LambdaClientContext) Base64Json() (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	if err := json.NewEncoder(encoder).Encode(lc); err != nil {
		return "", err
	}

	if err := encoder.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
