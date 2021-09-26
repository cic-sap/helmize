package zerolog

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Empty struct {
}

var levelMap = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"trace": zerolog.TraceLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

func setLevel() {
	envLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if v, ok := levelMap[envLevel]; ok {
		zerolog.SetGlobalLevel(v)
	}

}
func init() {

	setLevel()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	out.TimeFormat = time.RFC3339
	hostname, _ := os.Hostname()
	log.Logger = log.Output(out).
		With().
		Caller().
		//CallerWithSkipFrameCount(4).
		Stack().
		Int("pid", os.Getpid()).
		Str("process", filepath.Base(os.Args[0])).
		Str("hostname", hostname).
		Logger().
		//Hook(levelHook{}).
		Hook(goIDHook{})
}

type goIDHook struct{}

var rootId = GetGoid()

func (h goIDHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if rootId > 0 {
		e.Int64("goId", GetGoid())
	}
}

type levelHook struct{}

func (h levelHook) Run(_ *zerolog.Event, _ zerolog.Level, _ string) {
	setLevel()
}

func GetGoid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return int64(id)
}
