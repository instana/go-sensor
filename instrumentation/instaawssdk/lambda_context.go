package instaawssdk

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
