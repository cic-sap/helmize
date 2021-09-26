package utils

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"runtime"
	"testing"
	"time"
)

func TestRunShellOutput(t *testing.T) {

	log.Debug().Msgf("NumGoroutine:%v", runtime.NumGoroutine())
	time.Sleep(time.Second)
	w := bytes.NewBuffer(nil)

	s1, s2, err := RunShellOutput("ls", []string{"-alh"}, WithOutputStream(w))
	time.Sleep(time.Second)
	t.Log("\n############3\n")
	t.Log("err", err)
	t.Log("s1", s1, "s2", s2)
	log.Debug().Msgf("NumGoroutine:%v", runtime.NumGoroutine())
	log.Debug().Msg("all out:" + w.String())
}
