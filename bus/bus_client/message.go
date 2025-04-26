// file: buz/bus/bus_client/message.go

package bus_client

import (
	"fmt"

	"github.com/rskv-p/buzz/bus/bus_comm"
)

//-----------------------------------------
//  Message
//-----------------------------------------

// Message represents a message that can be published or processed in the bus.
type Message struct {
	Subject string            // The subject of the message
	Data    []byte            // The data of the message
	Headers map[string]string // Headers for the message
}

//-----------------------------------------
//  Message Creation
//-----------------------------------------

// NewMessage creates a new message with the specified subject and data.
func NewMessage(subject string, data []byte) *Message {
	bus_comm.Infof("Creating new message with subject: %s", subject)
	return &Message{
		Subject: subject,
		Data:    data,
		Headers: make(map[string]string), // Initialize headers as empty map
	}
}

//-----------------------------------------
//  Header Management
//-----------------------------------------

// SetHeader sets a header for the message.
func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
	bus_comm.Infof("Set header: key = %s, value = %s for message with subject: %s", key, value, m.Subject)
}

// GetHeader retrieves the value of a header by its key.
func (m *Message) GetHeader(key string) (string, error) {
	if value, exists := m.Headers[key]; exists {
		bus_comm.Infof("Retrieved header: key = %s, value = %s for message with subject: %s", key, value, m.Subject)
		return value, nil
	}
	bus_comm.Errorf("Header %s not found for message with subject: %s", key, m.Subject)
	return "", fmt.Errorf("header %s not found", key)
}

//-----------------------------------------
//  String Representation
//-----------------------------------------

// String returns a string representation of the message.
func (m *Message) String() string {
	messageStr := fmt.Sprintf("Subject: %s, Data: %s, Headers: %v", m.Subject, string(m.Data), m.Headers)
	bus_comm.Debugf("Message string representation: %s", messageStr)
	return messageStr
}
