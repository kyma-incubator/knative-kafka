package log

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger  *zap.Logger
	once    sync.Once
	watcher *fsnotify.Watcher
	cfg     zap.Config
)

const DefaultConfigFilePath = "/etc/knative-kafka/logging.json"

// Set Up Logger And File Watcher With Default Configuration File
func Logger() *zap.Logger {
	return LoggerWithConfig(DefaultConfigFilePath)
}

// Set Up Logger And File Watcher For Given Configuration File
func LoggerWithConfig(configFilePath string) *zap.Logger {
	once.Do(func() {
		newLogger, err := initLogger(configFilePath)
		if err != nil {
			panic(fmt.Errorf("Unable to configure logger, exiting: %v", err))
		}
		logger = newLogger
		initFileWatcher(configFilePath)
	})
	return logger
}

func TestLogger() *zap.Logger {
	once.Do(func() {
		logger, _ = zap.NewDevelopment()
	})
	return logger
}

// Load Log Configuration And Build Logger
// Notice that we need to load the configuration into a new struct
// and test the creation of the logger, before we apply the updates
// to the existing zap.Config. It is very important that the changes
// are applied to the existing zap.Config rather than using the new
// one so that child loggers get the configuration applied as well.
func initLogger(configFilePath string) (*zap.Logger, error) {
	// Open our jsonFile
	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println("Unable To Load Logging Configuration File", err)
		return nil, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// Check the configuration by trying to create a new logger
	var newCfg zap.Config
	if err := json.NewDecoder(jsonFile).Decode(&newCfg); err != nil {
		fmt.Println("Unable To Parse Logging Configuration File", err)
		return nil, err
	}
	newLogger, err := newCfg.Build()
	if err != nil {
		fmt.Println("Unable To Build Logger From Configuration", err)
		return nil, err
	}

	// Configuration is good, actually create the new config/logger
	jsonFile.Seek(0, 0)
	if err := json.NewDecoder(jsonFile).Decode(&cfg); err != nil {
		fmt.Println("Unable To Parse Logging Configuration File", err)
		return nil, err
	}
	newLogger, err = cfg.Build()

	return newLogger, err
}

// Initialize Watching Log Config File
func initFileWatcher(configFilePath string) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(fmt.Errorf("Unable To Setup Logging Config File Watcher: %v", err))
	}
	startWatchListener(configFilePath)
}

func onWatchEvent(event fsnotify.Event, configFilePath string, configFile string, realConfigFile string) {
	currentConfigFile, _ := filepath.EvalSymlinks(configFilePath)
	// we only care about the config file with the following cases:
	// 1 - if the config file was modified or created
	// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)
	const writeOrCreateMask = fsnotify.Write | fsnotify.Create
	if (filepath.Clean(event.Name) == configFile &&
		event.Op&writeOrCreateMask != 0) ||
		(currentConfigFile != "" && currentConfigFile != realConfigFile) {
		realConfigFile = currentConfigFile
		newLogger, err := initLogger(configFilePath)
		if err != nil {
			logger.Error("Unable To Update Logger Configuration, Keeping Existing Logger", zap.Error(err))
		} else {
			logger = newLogger
			logger.Info("Updated Logging Configuration")
		}
	}
}

// File Watch Listener, Based On Code From Viper
func startWatchListener(configFilePath string) {
	configFile := filepath.Clean(configFilePath)
	configDir, _ := filepath.Split(configFile)
	realConfigFile, _ := filepath.EvalSymlinks(configFilePath)
	watcher.Add(configDir)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok { // 'Events' channel is closed
					return
				}

				onWatchEvent(event, configFilePath, configFile, realConfigFile)

			case err, ok := <-watcher.Errors:
				if ok { // 'Errors' channel is not closed
					log.Printf("watcher error: %v\n", err)
				}
				return
			}
		}
	}()
}

// Cleanup The Various Parts Of Logging System
func Shutdown() {
	if watcher != nil {
		watcher.Close()
	}

	if logger != nil {
		logger.Sync()
	}

	once = sync.Once{}
}
