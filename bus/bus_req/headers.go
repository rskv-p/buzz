// file: buz/bus/bus_req/headers.go

package bus_req

import (
	"fmt"
	"sync"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_type"
)

var _ bus_type.IRequestHeaders = (*RequestHeaders)(nil) // Ensure RequestHeaders implements IRequestHeaders

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
		bus_comm.Infof("Successfully retrieved header for key '%s' with value '%s'", key, values[0])
		return values[0], nil
	}
	bus_comm.Warnf("Header not found for key '%s'", key)
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
	bus_comm.Infof("Header set for key '%s' with value '%s'", key, value)
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
		bus_comm.Infof("Header with key '%s' removed successfully", key)
		return nil
	}
	bus_comm.Warnf("Header not found for key '%s' to remove", key)
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
	bus_comm.Infof("Returning all headers")
	return headersCopy
}
