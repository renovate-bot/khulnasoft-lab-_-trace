package trace

import (
	"io"
	"testing"

	"github.com/jonboulle/clockwork"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type HooksSuite struct {
	suite.Suite
}

func TestHooks(t *testing.T) {
	suite.Run(t, new(HooksSuite))
}

func (s *HooksSuite) TestSafeForConcurrentAccess() {
	logger := log.New()
	logger.Out = io.Discard
	entry := logger.WithFields(log.Fields{"foo": "bar"})
	logger.Hooks.Add(&UDPHook{Clock: clockwork.NewFakeClock()})
	for i := 0; i < 3; i++ {
		go func(entry *log.Entry) {
			for i := 0; i < 1000; i++ {
				entry.Infof("test")
			}
		}(entry)
	}
}
