package api

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Just make sure that the server starts and terminates.
func TestAPISimple(t *testing.T) {
	tempLogDir := t.TempDir()
	defer os.RemoveAll(tempLogDir)

	log.Print(tempLogDir)

	s := New(&ServerConfig{
		ListenAddr:   ":0",
		AccessLogDir: tempLogDir,
	})
	assert.NotNil(t, s)
	s.Start()
	s.Shutdown()
}
