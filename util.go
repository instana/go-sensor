package instana

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	seededIDGen  = rand.New(rand.NewSource(time.Now().UnixNano()))
	seededIDLock sync.Mutex
)

func randomID() uint64 {
	seededIDLock.Lock()
	defer seededIDLock.Unlock()
	return uint64(seededIDGen.Int63())
}

func getCommandLine() (string, []string) {
	var cmdlinePath string = "/proc/" + strconv.Itoa(os.Getpid()) + "/cmdline"

	cmdline, err := ioutil.ReadFile(cmdlinePath)

	if err != nil {
		log.debug("No /proc.  Returning OS reported cmdline")
		return os.Args[0], os.Args[1:]
	}

	parts := strings.FieldsFunc(string(cmdline), func(c rune) bool {
		if c == '\u0000' {
			return true
		}
		return false
	})
	log.debug("cmdline says:", parts[0], parts[1:])
	return parts[0], parts[1:]
}
