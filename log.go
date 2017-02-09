package instana

import (
	"io/ioutil"
	"log"
)

// Logger is an instance of the StdLogger interface, used by the sensor to log relevant events.
// By default it is set to discard all log messages via ioutil.Discard.
// Since it is an exported global variable, it can be set to redirect to any desired output.
var Logger StdLogger = log.New(ioutil.Discard, "[instana]", log.LstdFlags)

// StdLogger declares methods for logging messages
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
