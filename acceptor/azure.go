// (c) Copyright IBM Corp. 2022

package acceptor

// NewAzurePluginPayload returns payload for the Azure plugin of Instana acceptor
func NewAzurePluginPayload(entityID string) PluginPayload {
	const pluginName = "com.instana.plugin.azure.functionapp"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
	}
}
