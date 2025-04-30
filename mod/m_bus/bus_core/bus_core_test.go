// file: buzz/mod/m_bus/bus_core/core_test.go

package bus_core_test

import (
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/buzz/mod/m_bus/bus_client"
	"github.com/rskv-p/buzz/mod/m_bus/bus_core"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
	"github.com/stretchr/testify/assert"
)

// TestRetrySubscribe tests retry subscription mechanism
func TestRetrySubscribe(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Create main bus with secret key
	bus := bus_core.NewBus("secret-key", nil, 5)
	x_log.Info("Bus created with secret key")

	// Create client using ClientFactory
	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)
	x_log.Info("Client 1 created and ready")

	// Register message handler
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler")
	x_log.Info("Handler registered for subject", "test.topic")

	// Subscribe client to topic with handler
	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client")
	x_log.Info("Client subscribed to", "test.topic", "with queue", "testQueue")

	// Retry publish message
	message := []byte("Test Message")
	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish")
	x_log.Info("Message", "Test Message", "published")

	// Wait for message processing using sync.WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			if receivedMessage != nil && string(receivedMessage) == "Test Message" {
				x_log.Info("Message received successfully")
				assert.Equal(t, "Test Message", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Timeout while waiting for message")
	case <-done:
		t.Log("Test completed successfully")
	}

	wg.Wait()
	x_log.Info("TestRetrySubscribe completed")
}

// TestRetryPublish tests retry publish mechanism
func TestRetryPublish(t *testing.T) {
	var receivedMessage []byte
	bus := bus_core.NewBus("secret-key", nil, 5)
	x_log.Info("Bus created with secret key")

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)
	x_log.Info("Client 1 created and ready")

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler")
	x_log.Info("Handler registered for subject", "test.topic")

	// Subscribe client to topic with handler
	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client")
	x_log.Info("Client subscribed to", "test.topic", "with queue", "testQueue")

	// Retry publish message
	message := []byte("Test Retry Message")
	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish")
	x_log.Info("Message", "Test Retry Message", "published")

	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			x_log.Info("Checking for received message...")
			if receivedMessage != nil && string(receivedMessage) == "Test Retry Message" {
				x_log.Info("Message received successfully")
				assert.Equal(t, "Test Retry Message", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		x_log.Error("Timeout while waiting for message")
		t.Fatal("Timeout while waiting for message")
	case <-done:
		t.Log("Test completed successfully")
	}

	wg.Wait()
	x_log.Info("TestRetryPublish completed")
}
