package instana

import (
	"time"
)

type EventData struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	// Duration in milliseconds
	Duration int `json:"duration"`
	// Severity with value of -1, 5, 10 : see type severity
	Severity int    `json:"severity"`
	Plugin   string `json:"plugin,omitempty"`
	ID       string `json:"id,omitempty"`
	Host     string `json:"host,omitempty"`
}

type severity int

//Severity values for events sent to the instana agent
const (
	SeverityChange   severity = -1
	SeverityWarning  severity = 5
	SeverityCritical severity = 10
)

const (
	ServicePlugin = "com.instana.forge.connection.http.logical.LogicalWebApp"
	ServiceHost   = ""
)

//SendDefaultServiceEvent sends a default event which already contains the service and host
func SendDefaultServiceEvent(title string, text string, sev severity, duration time.Duration) {
	SendServiceEvent(sensor.serviceName, title, text, sev, duration)
}

func SendServiceEvent(service string, title string, text string, sev severity, duration time.Duration) {
	sendEvent(&EventData{
		Title:    title,
		Text:     text,
		Severity: int(sev),
		Plugin:   ServicePlugin,
		ID:       service,
		Host:     ServiceHost,
		Duration: int(duration / time.Millisecond),
	})
}

func SendHostEvent(title string, text string, sev severity, duration time.Duration) {
	sendEvent(&EventData{
		Title:    title,
		Text:     text,
		Duration: int(duration / time.Millisecond),
		Severity: int(sev),
	})
}

func sendEvent(event *EventData) {

	log.debug(event)

	//we do fire & forget here, because the whole pid dance isn't necessary to send events
	go sensor.agent.request(sensor.agent.makeURL(agentEventURL), "POST", event)
}
