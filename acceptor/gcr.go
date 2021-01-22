// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor

// GCRServiceRevisionInstanceData is a representation of a Google Cloud Run service revision instance
// for com.instana.plugin.gcp.run.revision.instance plugin
type GCRServiceRevisionInstanceData struct {
	Runtime          string `json:"runtime,omitempty"`
	Region           string `json:"region"`
	Service          string `json:"service"`
	Configuration    string `json:"configuration,omitempty"`
	Revision         string `json:"revision"`
	InstanceID       string `json:"instanceId"`
	Port             string `json:"port,omitempty"`
	NumericProjectID int    `json:"numericProjectId"`
	ProjectID        string `json:"projectId,omitempty"`
}

// NewGCRServiceRevisionInstancePluginPayload returns payload for the GCR service revision instance
// plugin of Instana acceptor
func NewGCRServiceRevisionInstancePluginPayload(entityID string, data GCRServiceRevisionInstanceData) PluginPayload {
	const pluginName = "com.instana.plugin.gcp.run.revision.instance"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
		Data:     data,
	}
}
