package utils

import (
	"os"
	"time"

	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
)

func RestartServer() {
	go func() {
		log.Debug("Restarting server in 0.5 seconds")
		time.Sleep(1 * time.Second)
		log.Debug("Restarting NOW")
		types.Quit <- os.Kill
	}()

}
