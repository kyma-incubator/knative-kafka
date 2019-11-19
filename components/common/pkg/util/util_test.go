package util

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"syscall"
	"testing"
	"time"
)

// Test The WaitForSignal Functionality
func TestWaitForSignal(t *testing.T) {

	// Logger Reference
	var testLogger = getLogger(t)

	// Create A Completion Channel
	doneChan := make(chan struct{})

	// Start Async Go Routing
	go func() {

		// Perform The Test - Wait For SIGINT
		testLogger.Info("Waiting For Signal")
		WaitForSignal(testLogger, syscall.SIGINT)

		// Close The Completion Channel
		close(doneChan)
	}()

	// Wait A Short Bit To Let Async Function Start
	testLogger.Info("Sleeping")
	time.Sleep(500 * time.Millisecond)

	// Issue The Signal
	testLogger.Info("Sending Signal")
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	assert.Nil(t, err)

	// Block Waiting For Completion
	<-doneChan

	// Done!  (If The Test Completes (Doesn't Hang) Then It Was Successful)
	testLogger.Info("Done!")
}

// Initialize The Logger - Fatal Exit Upon Error
func getLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewProduction() // For Now Just Use The Default Zap Production Logger
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	return logger
}
