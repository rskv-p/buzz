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

// TestRetrySubscribe tests retry subscription mechanism for both IPv4 and IPv6 addresses
func TestRetrySubscribe(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Test with IPv4 address
	busIPv4 := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8080)   // Use local IPv4 address
	assert.NoError(t, busIPv4.Start(), "Failed to start bus (IPv4)") // Start the bus server
	x_log.Info("Bus created with secret key (IPv4)")

	// Create client using ClientFactory for IPv4
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

	// Retry publish message
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
	busIPv6 := bus_core.NewBus("secret-key", 5, "2001:db8::1", 8080) // Use a local IPv6 address
	assert.NoError(t, busIPv6.Start(), "Failed to start bus (IPv6)") // Start the bus server
	x_log.Info("Bus created with secret key (IPv6)")

	// Create client using ClientFactory for IPv6
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

	// Retry publish message
	err = clientIPv6.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish (IPv6)")
	x_log.Info("Message", "Test Message (IPv6)", "published")

	// Wait for message processing using sync.WaitGroup for IPv6
	wg.Add(1)
	done = make(chan struct{})

	go func() {
		defer wg.Done()
		for {
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

	// Stop the bus after the tests
	assert.NoError(t, busIPv4.Stop(), "Failed to stop bus (IPv4)")
	assert.NoError(t, busIPv6.Stop(), "Failed to stop bus (IPv6)")
}

// TestRetryPublish tests retry publish mechanism for both IPv4 and IPv6 addresses
func TestRetryPublish(t *testing.T) {
	var receivedMessage []byte

	// Test with IPv4 address
	busIPv4 := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8080) // Use local IPv4 address
	x_log.Info("Bus created with secret key (IPv4)")

	clientIPv4 := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8080)
	})(1, busIPv4)
	x_log.Info("Client 1 created and ready (IPv4)")

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

	// Retry publish message
	message := []byte("Test Retry Message")
	err = clientIPv4.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish (IPv4)")
	x_log.Info("Message", "Test Retry Message", "published (IPv4)")

	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			x_log.Info("Checking for received message...")
			if receivedMessage != nil && string(receivedMessage) == "Test Retry Message" {
				x_log.Info("Message received successfully (IPv4)")
				assert.Equal(t, "Test Retry Message", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		x_log.Error("Timeout while waiting for message (IPv4)")
		t.Fatal("Timeout while waiting for message (IPv4)")
	case <-done:
		t.Log("Test completed successfully (IPv4)")
	}

	wg.Wait()
	x_log.Info("TestRetryPublish completed (IPv4)")

	// Test with IPv6 address
	busIPv6 := bus_core.NewBus("secret-key", 5, "2001:db8::1", 8080) // Use a local IPv6 address
	x_log.Info("Bus created with secret key (IPv6)")

	clientIPv6 := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "2001:db8::1", 8080)
	})(2, busIPv6)
	x_log.Info("Client 2 created and ready (IPv6)")

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

	// Retry publish message
	err = clientIPv6.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish (IPv6)")
	x_log.Info("Message", "Test Retry Message", "published (IPv6)")

	wg.Add(1)
	done = make(chan struct{})

	go func() {
		defer wg.Done()
		for {
			x_log.Info("Checking for received message...")
			if receivedMessage != nil && string(receivedMessage) == "Test Retry Message" {
				x_log.Info("Message received successfully (IPv6)")
				assert.Equal(t, "Test Retry Message", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		x_log.Error("Timeout while waiting for message (IPv6)")
		t.Fatal("Timeout while waiting for message (IPv6)")
	case <-done:
		t.Log("Test completed successfully (IPv6)")
	}

	wg.Wait()
	x_log.Info("TestRetryPublish completed (IPv6)")
}
