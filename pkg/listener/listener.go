// Package listener defines the listener interface.
package listener

// Listener interface
type Listener interface {
	Start() error
	Shutdown()
	WaitForShutdown()
}
