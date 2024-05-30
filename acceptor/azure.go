// (c) Copyright IBM Corp. 2022

package acceptor

// NewAzurePluginPayload returns payload for the Azure plugin of Instana acceptor
func NewAzurePluginPayload(entityID, pluginName string) PluginPayload {
	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
	}
}
