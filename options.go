package instana

type Options struct {
	Service                     string
	AgentHost                   string
	AgentPort                   int
	MaxBufferedSpans            int
	ForceTransmissionStartingAt int
	LogLevel                    int
	TestMode                    bool
}
