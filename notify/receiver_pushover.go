package notify

import (
	"fmt"
	"github.com/gregdel/pushover"
	log "github.com/sirupsen/logrus"
	"notdeadyet/config"
	"time"
)

type PushoverReceiver struct {
	config *config.PushoverReceiverConfig

	pushover  *pushover.Pushover
	recipient *pushover.Recipient

	logger *log.Entry
}

func NewPushoverReceiver(cfg *config.PushoverReceiverConfig) (*PushoverReceiver, error) {
	return &PushoverReceiver{
		config:    cfg,
		pushover:  pushover.New(cfg.Token),
		recipient: pushover.NewRecipient(cfg.UserKey),
		logger:    log.WithField("receiver", cfg.Name),
	}, nil
}

func (p *PushoverReceiver) Config() *config.ReceiverConfig {
	return &p.config.ReceiverConfig
}

func (p *PushoverReceiver) NotifyAppDown(appName string, downtime time.Duration) error {
	return p.sendMessage(
		fmt.Sprintf("App \"%s\" has not sent a live sign since %v. It's probably dead.", appName, downtime),
		fmt.Sprintf("%s has died.", appName),
	)
}

func (p *PushoverReceiver) NotifyAppBack(appName string) error {
	return p.sendMessage(
		fmt.Sprintf("App \"%s\" has reappeared after being dead.", appName),
		fmt.Sprintf("%s is back!", appName),
	)
}

func (p *PushoverReceiver) sendMessage(message string, title string) error {
	msg := pushover.NewMessageWithTitle(message, title)
	msg.Priority = p.config.Priority
	_, err := p.pushover.SendMessage(msg, p.recipient)
	if err != nil {
		return fmt.Errorf("sending message failed: %w", err)
	}
	return nil
}
