package bus_client_test

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

// TestBusClient tests the creation of a BusClient and its message handling
func TestBusClient(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Test with IPv4 address
	t.Log("Creating bus with secret key 'secret-key' (IPv4)")

	// Create the main bus with IPv4 address
	busIPv4 := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8080) // Use local IPv4 address
	x_log.Info("Bus created with secret key (IPv4)")

	// Create the client using IPv4 address
	clientIPv4 := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8080)
	})(1, busIPv4)
	x_log.Info("Client 1 created and ready (IPv4)")

	// Register message handler for IPv4
	err := clientIPv4.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler (IPv4)")
	x_log.Info("Handler registered for subject", "test.topic")

	// Subscribe client to topic with handler
	err = clientIPv4.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client (IPv4)")
	x_log.Info("Client subscribed to", "test.topic", "with queue", "testQueue")

	// Retry publish message to IPv4 address
	message := []byte("Test Message (IPv4)")
	err = clientIPv4.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish (IPv4)")
	x_log.Info("Message", "Test Message (IPv4)", "published")

	// Wait for message processing using sync.WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			if receivedMessage != nil && string(receivedMessage) == "Test Message (IPv4)" {
				x_log.Info("Message received successfully (IPv4)")
				assert.Equal(t, "Test Message (IPv4)", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Timeout while waiting for message (IPv4)")
	case <-done:
		t.Log("Test completed successfully (IPv4)")
	}

	wg.Wait()
	x_log.Info("TestRetrySubscribe completed (IPv4)")

	// Test with IPv6 address
	t.Log("Creating bus with secret key 'secret-key' (IPv6)")

	// Create the main bus with IPv6 address
	busIPv6 := bus_core.NewBus("secret-key", 5, "2001:db8::1", 8080) // Use a local IPv6 address
	x_log.Info("Bus created with secret key (IPv6)")

	// Create the client using IPv6 address
	clientIPv6 := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "2001:db8::1", 8080)
	})(2, busIPv6)
	x_log.Info("Client 2 created and ready (IPv6)")

	// Register message handler for IPv6
	err = clientIPv6.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler (IPv6)")
	x_log.Info("Handler registered for subject", "test.topic")

	// Subscribe client to topic with handler
	err = clientIPv6.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client (IPv6)")
	x_log.Info("Client subscribed to", "test.topic", "with queue", "testQueue")

	// Retry publish message to IPv6 address
	err = clientIPv6.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish (IPv6)")
	x_log.Info("Message", "Test Message (IPv6)", "published")

	wg.Add(1)
	done = make(chan struct{})

	go func() {
		defer wg.Done()
		for {
			x_log.Info("Checking for received message...")
			if receivedMessage != nil && string(receivedMessage) == "Test Message (IPv6)" {
				x_log.Info("Message received successfully (IPv6)")
				assert.Equal(t, "Test Message (IPv6)", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Timeout while waiting for message (IPv6)")
	case <-done:
		t.Log("Test completed successfully (IPv6)")
	}

	wg.Wait()
	x_log.Info("TestRetrySubscribe completed (IPv6)")
}

// TestDuplicateSubscription checks if the client can subscribe to the same topic and queue multiple times.
func TestDuplicateSubscription(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Create the main bus with a secret key, host, and port (IPv4 for this test)
	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8080) // Use local IPv4 address

	// Create the client through ClientFactory
	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8080)
	})(1, bus)

	// Register the message handler for "test.topic"
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register message handler")

	// Subscribe the client to the "test.topic" and "testQueue" with a handler
	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client to 'test.topic' with 'testQueue'")

	// Attempt to subscribe to the same topic and queue again with a handler
	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client again to the same topic and queue")

	// Publish a message to the topic
	message := []byte("Test Message")
	err = client.PublishMessage("test.topic", message, "testQueue")
	assert.NoError(t, err, "Failed to publish message")

	// Wait for the message to be received
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for receivedMessage == nil || string(receivedMessage) != "Test Message" {
			time.Sleep(10 * time.Millisecond)
		}
		assert.Equal(t, "Test Message", string(receivedMessage), "Message was not received correctly")
	}()

	wg.Wait()
}

func TestMessageProcessing(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Create the main bus with a secret key and provide host and port (IPv4 address)
	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8080)

	// Create the client through ClientFactory, providing host and port as arguments
	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8080)
	})(1, bus)

	// Register the message handler for "test.topic"
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register message handler")

	// Subscribe the client to the "test.topic" and "testQueue" with a handler
	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to subscribe client to 'test.topic' with 'testQueue'")

	// Publish a message to the topic
	message := []byte("Test Processing Message")
	err = client.PublishMessage("test.topic", message, "testQueue")
	assert.NoError(t, err, "Failed to publish message")

	// Wait for the message to be received
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for receivedMessage == nil || string(receivedMessage) != "Test Processing Message" {
			time.Sleep(10 * time.Millisecond)
		}
		assert.Equal(t, "Test Processing Message", string(receivedMessage), "Message was not received correctly")
	}()

	wg.Wait()
}

func TestInvalidTopic(t *testing.T) {
	// Create the main bus with a secret key and provide host and port (IPv4 address)
	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8080)

	// Create the client through ClientFactory, providing host and port as arguments
	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8080)
	})(1, bus)

	// Try subscribing to a non-existing topic (invalid topic)
	err := client.Subscribe("invalid.topic", "testQueue", func(subject string, msg []byte) {
		// Do nothing, we expect failure
	})
	assert.Error(t, err, "Subscription to invalid topic should fail")

	// Try publishing to a non-existing topic
	message := []byte("Test Message to Invalid Topic")
	err = client.PublishMessage("invalid.topic", message, "testQueue")
	assert.Error(t, err, "Publishing to invalid topic should fail")
}
