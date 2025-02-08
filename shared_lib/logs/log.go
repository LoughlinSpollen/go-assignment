package logs

import (
	env "assignment/lib/shared_lib/env"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
)

func debugTags() {
	log.SetLevel(log.TraceLevel)
	log.SetOutput(os.Stdout)
}

func releaseTags() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)
}

// add OpenTelemetry hook
func openTelemetry() {
	log.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
		log.InfoLevel,
	)))
}

func Init() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	if env.IsLocalDev() {
		debugTags()
	} else {
		releaseTags()
		openTelemetry()
	}
}
