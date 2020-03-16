package log

// TODO - REMOVE OLD !!!
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/stretchr/testify/assert"
//	"go.uber.org/zap"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//	"testing"
//	"time"
//)
//
//func TestLoggerWithConfig(t *testing.T) {
//	tmpFile := setupTempLoggerConfig(t)
//	defer Shutdown()
//	defer os.Remove(tmpFile.Name())
//
//	logger := LoggerWithConfig(tmpFile.Name())
//
//	assert.NotNil(t, logger, "Logger Should Not Null")
//	assert.True(t, logger.Core().Enabled(zap.DebugLevel), "Generated Logger Should Support DEBUG Level")
//}
//
//func Test_initFileWatcher(t *testing.T) {
//	tmpFile := setupTempLoggerConfig(t)
//	defer Shutdown()
//	defer os.Remove(tmpFile.Name())
//
//	fmt.Println(tmpFile.Name())
//	configFile := filepath.Clean(tmpFile.Name())
//	configDir, _ := filepath.Split(configFile)
//
//	initFileWatcher(tmpFile.Name())
//	assert.NotNil(t, watcher)
//	assert.NoError(t, watcher.Remove(configDir))
//}
//
//func Test_ConfigDynamicallyChanged(t *testing.T) {
//
//	tmpFile := setupTempLoggerConfig(t)
//	defer Shutdown()
//	defer os.Remove(tmpFile.Name())
//
//	logger := LoggerWithConfig(tmpFile.Name())
//	assert.True(t, logger.Core().Enabled(zap.DebugLevel), "Initial Logger Should Support DEBUG Level")
//
//	updateTempLoggerConfig(t, tmpFile)
//	time.Sleep(1 * time.Second)
//	assert.False(t, logger.Core().Enabled(zap.DebugLevel), "Updated Logger Should Not Support DEBUG Level")
//}
//
//func Test_InvalidDynamicConfig(t *testing.T) {
//
//	tmpFile := setupTempLoggerConfig(t)
//	defer Shutdown()
//	defer os.Remove(tmpFile.Name())
//
//	logger := LoggerWithConfig(tmpFile.Name())
//	assert.True(t, logger.Core().Enabled(zap.DebugLevel), "Initial Logger Should Support DEBUG Level")
//
//	updateInvalidTempLoggerConfig(t, tmpFile)
//	time.Sleep(1 * time.Second)
//	assert.True(t, logger.Core().Enabled(zap.DebugLevel), "Logger Should Still DEBUG Level")
//}
//
//const initialTestLoggingConfig = `
//{
//  "level": "debug",
//  "development": false,
//  "outputPaths": [
//    "stdout"
//  ],
//  "errorOutputPaths": [
//    "stderr"
//  ],
//  "encoding": "json",
//  "encoderConfig": {
//    "timeKey": "ts",
//    "levelKey": "level",
//    "nameKey": "logger",
//    "callerKey": "caller",
//    "messageKey": "msg",
//    "stacktraceKey": "stacktrace",
//    "lineEnding": "",
//    "levelEncoder": "",
//    "timeEncoder": "iso8601",
//    "durationEncoder": "",
//    "callerEncoder": ""
//  }
//}
//`
//
//func setupTempLoggerConfig(t *testing.T) *os.File {
//
//	tmpFile, err := ioutil.TempFile(os.TempDir(), "logcfg-")
//	assert.Nil(t, err, "Cannot Create Temporary File")
//
//	_, err = tmpFile.Write([]byte(initialTestLoggingConfig))
//	assert.Nil(t, err, "Failed To Write To Tempoary File")
//
//	err = tmpFile.Close()
//	assert.Nil(t, err, "Failed To Close Temporary File")
//
//	return tmpFile
//}
//
//func updateTempLoggerConfig(t *testing.T, tmpFile *os.File) {
//	data := make(map[string]interface{})
//	err := json.Unmarshal([]byte(initialTestLoggingConfig), &data)
//	assert.Nil(t, err, "Unable To Unmarshal JSON")
//
//	data["level"] = "error"
//
//	dataArray, err := json.Marshal(data)
//	assert.Nil(t, err, "Unable To Marshal JSON")
//
//	tmpFile, err = os.OpenFile(tmpFile.Name(), os.O_WRONLY, os.ModePerm)
//	assert.Nil(t, err, "Unable To Open File")
//
//	_, err = tmpFile.Write(dataArray)
//	assert.Nil(t, err, "Failed To Write To Temporary File")
//
//	err = tmpFile.Close()
//	assert.Nil(t, err, "Unable To Close Temporary File")
//}
//
//func updateInvalidTempLoggerConfig(t *testing.T, tmpFile *os.File) {
//	data := make(map[string]interface{})
//	err := json.Unmarshal([]byte(initialTestLoggingConfig), &data)
//	assert.Nil(t, err, "Unable To Unmarshal JSON")
//
//	data["level"] = "error"
//	data["encoding"] = "foobar"
//
//	dataArray, err := json.Marshal(data)
//	assert.Nil(t, err, "Unable To Marshal JSON")
//
//	tmpFile, err = os.OpenFile(tmpFile.Name(), os.O_WRONLY, os.ModePerm)
//	assert.Nil(t, err, "Unable To Open File")
//
//	_, err = tmpFile.Write(dataArray)
//	assert.Nil(t, err, "Failed To Write To Temporary File")
//
//	err = tmpFile.Close()
//	assert.Nil(t, err, "Unable To Close Temporary File")
//}
