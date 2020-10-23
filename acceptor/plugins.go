// Package acceptor provides marshaling structs for Instana serverless acceptor API
package acceptor

// PluginPayload represents the Instana acceptor message envelope containing plugin
// name and entity ID
type PluginPayload struct {
	Name     string      `json:"name"`
	EntityID string      `json:"entityId"`
	Data     interface{} `json:"data"`
}
