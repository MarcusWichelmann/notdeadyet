package main

import (
	"crypto/subtle"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ginlogrus "github.com/toorop/gin-logrus"
	"net/http"
	"notdeadyet/config"
	"notdeadyet/notify"
	"notdeadyet/watching"
	"os"
)

type Options struct {
	ConfigFile string `short:"c" long:"config-file" description:"The config file to load, can be a .toml, .yml or .json"`
	LogLevel   string `short:"l" long:"log-level" description:"The log level" choice:"debug" choice:"info" choice:"warn" default:"info"`
	LogJson    bool   `short:"j" long:"log-json" description:"Write logs as json"`
}

var (
	options     Options
	flagsParser = flags.NewParser(&options, flags.Default)

	cfg config.Config

	receivers []notify.Receiver
	watchers  []*watching.Watcher
)

func main() {
	// Parse CLI options.
	if _, err := flagsParser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		} else {
			log.Fatal(err)
		}
	}

	// Configure logging.
	logLevel, err := log.ParseLevel(options.LogLevel)
	if err != nil {
		log.WithError(err).Fatal("Unknown log level.")
	}
	log.SetLevel(logLevel)
	if options.LogJson {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Specify configuration file location.
	if options.ConfigFile != "" {
		viper.SetConfigFile(options.ConfigFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/notdeadyet/")
		viper.AddConfigPath(".")
	}

	// Read configuration file.
	log.Info("Parsing configuration...")
	config.SetDefaults(viper.GetViper())
	if err := viper.ReadInConfig(); err != nil {
		log.WithField("File", viper.ConfigFileUsed()).WithError(err).Fatal("Cannot read config file.")
	}

	// Unmarshal configuration.
	if err := viper.Unmarshal(&cfg); err != nil {
		log.WithField("File", viper.ConfigFileUsed()).WithError(err).Fatal("Error unmarshalling config file.")
	}
	log.WithField("File", viper.ConfigFileUsed()).Debug("Configuration parsed.")

	// Create receivers.
	initReceivers()

	// Create and start watchers.
	initWatchers()

	// Configure gin mode.
	if log.IsLevelEnabled(log.DebugLevel) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configure the HTTP server engine.
	g := gin.New()
	g.Use(ginlogrus.Logger(log.StandardLogger()), gin.Recovery())
	g.GET("/", handleIndex)
	g.GET("/im-alive/:token", handleLiveSign)
	g.POST("/im-alive/:token", handleLiveSign)

	// Run HTTP server.
	log.Info("Listening...")
	if err := g.Run(cfg.Listen); err != nil {
		log.WithError(err).Fatal("Starting HTTP server failed.")
	}
}

func initReceivers() {
	for _, rc := range cfg.Receivers.PushoverReceivers {
		r, err := notify.NewPushoverReceiver(&rc)
		if err != nil {
			log.WithField("receiver", rc.Name).WithError(err).Fatal("Cannot create pushover receiver.")
		}
		receivers = append(receivers, r)
	}
}

func initWatchers() {
	for _, ac := range cfg.Apps {
		// Filter receivers.
		filteredReceivers := make([]notify.Receiver, 0)
		for _, receiverName := range ac.NotificationReceivers {
			var receiver notify.Receiver
			for _, r := range receivers {
				if r.Config().Name == receiverName {
					receiver = r
					break
				}
			}
			if receiver == nil {
				log.WithField("app", ac.Name).Fatalf("Receiver \"%s\" does not exist.", receiverName)
			}
			filteredReceivers = append(filteredReceivers, receiver)
		}

		// Create watcher.
		w, err := watching.NewWatcher(ac, filteredReceivers)
		if err != nil {
			log.WithField("app", ac.Name).WithError(err).Fatal("Cannot create app watcher.")
		}
		w.Start()
		watchers = append(watchers, w)
	}
}

func handleIndex(c *gin.Context) {
	c.String(http.StatusOK, "Not Dead Yet? - The dead man's switch monitoring daemon.")
}

func handleLiveSign(c *gin.Context) {
	// Find the app.
	token := c.Param("token")
	var foundWatcher *watching.Watcher
	for _, w := range watchers {
		// Simple protection against brute force attacks.
		if subtle.ConstantTimeCompare([]byte(w.App.Token), []byte(token)) == 1 {
			foundWatcher = w
			break
		}
	}
	if foundWatcher == nil {
		c.String(http.StatusNotFound, "Unknown token.")
		return
	}

	foundWatcher.HandleLiveSign()

	c.String(http.StatusOK, "Got it. Waiting 'till you die.")
}
