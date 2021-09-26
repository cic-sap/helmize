package zerolog

import (
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"testing"
)

func TestGoId(t *testing.T) {

	os.Setenv("LOG_LEVEL", "warn")
	setLevel()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			log.Info().Str("MSG", "33").Msg("TEST LOG")
			wg.Done()
		}()
	}

	wg.Wait()
	log.Info().Str("MSG", "33").Msg("TEST LOG")
	log.Warn().Str("MSG", "33").Msg("TEST warn LOG")
	log.Debug().Str("MSG", "33").Msg("TEST debug LOG")
}
