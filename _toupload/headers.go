// file: buzz/mod/m_bus/bus_req/headers.go

package bus_req

import (
	"fmt"
	"sync"

	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

var _ typ.IRequestHeaders = (*RequestHeaders)(nil) // Ensure RequestHeaders implements IRequestHeaders

//-------------------------------------------------
// RequestHeaders - Represents HTTP-like headers for the request
//-------------------------------------------------

// RequestHeaders represents HTTP-like headers for the request.
type RequestHeaders struct {
	Headers map[string][]string // Map to store header keys and values
	mu      sync.RWMutex        // Mutex for thread safety in concurrent access
}

//-------------------------------------------------
// Get - Retrieves the first value for the given key from the headers
//-------------------------------------------------

// Get retrieves the first value for the given key from the headers.
func (h *RequestHeaders) Get(key string) (string, error) {
	h.mu.RLock() // Lock for reading
	defer h.mu.RUnlock()

	// Check if the header exists and return the first value
	if values, exists := h.Headers[key]; exists && len(values) > 0 {
		x_log.Info("Successfully retrieved header", "key", key, "value", values[0])
		return values[0], nil
	}
	x_log.Warn("Header not found", "key", key)
	return "", fmt.Errorf("header not found for key: %s", key)
}

//-------------------------------------------------
// Set - Sets the value for a header key
//-------------------------------------------------

// Set sets the value for a header key.
func (h *RequestHeaders) Set(key, value string) error {
	h.mu.Lock() // Lock for writing
	defer h.mu.Unlock()

	// Initialize the map if it's nil
	if h.Headers == nil {
		h.Headers = make(map[string][]string)
	}

	// Set the header value
	h.Headers[key] = []string{value}
	x_log.Info("Header set", "key", key, "value", value)
	return nil
}

//-------------------------------------------------
// Remove - Removes a header key-value pair
//-------------------------------------------------

// Remove removes a header key-value pair.
func (h *RequestHeaders) Remove(key string) error {
	h.mu.Lock() // Lock for writing
	defer h.mu.Unlock()

	// Check if the header exists, then remove it
	if _, exists := h.Headers[key]; exists {
		delete(h.Headers, key)
		x_log.Info("Header removed", "key", key)
		return nil
	}
	x_log.Warn("Header not found to remove", "key", key)
	return fmt.Errorf("header not found for key: %s", key)
}

//-------------------------------------------------
// GetAll - Returns all headers as a map
//-------------------------------------------------

// GetAll returns all headers as a map.
func (h *RequestHeaders) GetAll() map[string][]string {
	h.mu.RLock() // Lock for reading
	defer h.mu.RUnlock()

	// Return a copy of the headers to prevent external modifications
	headersCopy := make(map[string][]string)
	for k, v := range h.Headers {
		headersCopy[k] = v
	}
	x_log.Info("Returning all headers")
	return headersCopy
}
