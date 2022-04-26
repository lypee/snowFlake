package base

import "testing"

func TestDebugF(t *testing.T) {
	logger := DefaultLogger()
	s := "127.0.0.1"
	logger.DebugF("server %s error!", s)
}