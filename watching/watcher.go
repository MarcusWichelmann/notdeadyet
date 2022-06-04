package watching

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"notdeadyet/config"
	"notdeadyet/notify"
	"sync"
	"time"
)

type Watcher struct {
	App       *config.AppConfig
	Receivers []notify.Receiver

	timeout        time.Duration
	repeatInterval time.Duration

	mutex sync.Mutex

	lastLiveSign time.Time
	liveSign     chan struct{}

	timeoutExceeded bool

	logger *log.Entry
}

func NewWatcher(app *config.AppConfig, receivers []notify.Receiver) (*Watcher, error) {
	// Parse durations.
	timeout, err := time.ParseDuration(app.Timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot parse timeout %s: %w", app.Timeout, err)
	}
	repeatInterval, err := time.ParseDuration(app.RepeatInterval)
	if err != nil {
		return nil, fmt.Errorf("cannot parse repeat interval %s: %w", app.RepeatInterval, err)
	}

	return &Watcher{
		App:            app,
		Receivers:      receivers,
		timeout:        timeout,
		repeatInterval: repeatInterval,
		lastLiveSign:   time.Now().UTC(),
		liveSign:       make(chan struct{}),
		logger:         log.WithField("app", app.Name),
	}, nil
}

func (w *Watcher) Start() {
	w.logger.Info("Watching app...")
	go w.watch()
}

func (w *Watcher) HandleLiveSign() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.lastLiveSign = time.Now().UTC()
	w.liveSign <- struct{}{}

	if w.timeoutExceeded {
		w.logger.Info("App has risen from the dead, welcome back!")
		w.timeoutExceeded = false
		go w.notifyAppBack()
	}
}

func (w *Watcher) watch() {
	for {
		// Create a new timer on each iteration.
		// The duration depends on whether the timeout already exceeded
		// and we just wait to send repeated alerts or not.
		w.mutex.Lock()
		var nextDuration time.Duration
		if w.timeoutExceeded {
			nextDuration = w.repeatInterval
		} else {
			nextDuration = w.timeout
		}
		w.mutex.Unlock()
		tt := time.NewTimer(nextDuration)

		select {
		case <-tt.C:
			if !w.timeoutExceeded {
				w.logger.Info("Timeout reached. The app has probably died.")

				w.mutex.Lock()
				w.timeoutExceeded = true
				w.mutex.Unlock()
			}
			go w.notifyAppDown()
		case <-w.liveSign:
			// Cancel the timer.
			if !tt.Stop() {
				<-tt.C
			}
		}
	}
}

func (w *Watcher) notifyAppDown() {
	downtime := time.Now().UTC().Sub(w.lastLiveSign)
	w.logger.WithField("downtime", downtime).Info("Notifying receivers, that the app is (still) down.")
	for _, r := range w.Receivers {
		if err := r.NotifyAppDown(w.App.Name, downtime); err != nil {
			w.logger.WithError(err).Error("Sending app down notification failed.")
		}
	}
}

func (w *Watcher) notifyAppBack() {
	w.logger.Info("Notifying receivers, that the app is back.")
	for _, r := range w.Receivers {
		if err := r.NotifyAppBack(w.App.Name); err != nil {
			w.logger.WithError(err).Error("Sending app back notification failed.")
		}
	}
}
