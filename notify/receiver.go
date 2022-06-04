package notify

import (
	"notdeadyet/config"
	"time"
)

type Receiver interface {
	Config() *config.ReceiverConfig
	NotifyAppDown(appName string, downtime time.Duration) error
	NotifyAppBack(appName string) error
}
