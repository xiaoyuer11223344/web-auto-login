package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"xiaoyu/cmd"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableColors:   false,
		DisableSorting:  true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	cmd.Execute()
}
