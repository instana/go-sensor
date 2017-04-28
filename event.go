package instana

type EventData struct {
	Title    string `json:"title"`
	Text     string `json:"text"`
	Duration int    `json:"duration"`
	Severity int    `json:"severity"`
	Plugin   string `json:"plugin,omitempty"`
	ID       string `json:"id,omitempty"`
	Host     string `json:"host,omitempty"`
}

const (
	ServicePlugin = "com.instana.forge.connection.http.logical.LogicalWebApp"
	ServiceHost   = ""
)

func SendDefaultServiceEvent(title string, text string, severity int) {
	SendServiceEvent(sensor.serviceName, title, text, severity)
}

func SendServiceEvent(service string, title string, text string, severity int) {
	sendEvent(&EventData{
		Title:    title,
		Text:     text,
		Duration: 0,
		Severity: severity,
		Plugin:   ServicePlugin,
		ID:       service,
		Host:     ServiceHost})
}

func SendHostEvent(title string, text string, severity int) {
	sendEvent(&EventData{
		Title:    title,
		Text:     text,
		Duration: 0,
		Severity: severity})
}

func sendEvent(event *EventData) {

	log.debug(event)

	//we do fire & forget here, because the whole pid dance isn't necessary to send events
	go sensor.agent.request(sensor.agent.makeURL(agentEventURL), "POST", event)
}
