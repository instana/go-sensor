package instaawssdk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// LambdaClientContextClientApplication represent client application specific data part of the LambdaClientContext.
type LambdaClientContextClientApplication struct {
	InstallationID string `json:"installation_id"`
	AppTitle       string `json:"app_title"`
	AppVersionCode string `json:"app_version_code"`
	AppPackageName string `json:"app_package_name"`
}

// LambdaClientContext represents ClientContext from the AWS Invoke call https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax.
type LambdaClientContext struct {
	Client LambdaClientContextClientApplication
	Env    map[string]string `json:"env"`
	Custom map[string]string `json:"custom"`
}

// Base64JSON marshal LambdaClientContext to JSON and returns it as the base64 encoded string or error if any occurs.
func (lc *LambdaClientContext) Base64JSON() (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	if err := json.NewEncoder(encoder).Encode(*lc); err != nil {
		return "", fmt.Errorf("lambda client context encoder encode: %v", err.Error())
	}

	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("lambda client context encoder close: %v", err.Error())
	}

	return buf.String(), nil
}

// NewLambdaClientContextFromBase64EncodedJSON creates LambdaClientContext from the base64 encoded JSON or
// error if there is decoding error.
func NewLambdaClientContextFromBase64EncodedJSON(data string) (LambdaClientContext, error) {
	reader := strings.NewReader(data)
	decoder := base64.NewDecoder(base64.StdEncoding, reader)

	res := LambdaClientContext{}
	if err := json.NewDecoder(decoder).Decode(&res); err != nil {
		return res, fmt.Errorf("can't decode lambda client context: %v", err.Error())
	}

	return res, nil
}
