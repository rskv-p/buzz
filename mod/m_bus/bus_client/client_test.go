// file:buzz/mod/m_bus/bus_client/client_test.go

package bus_client_test

import (
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/buzz/mod/m_bus/bus_client"
	"github.com/rskv-p/buzz/mod/m_bus/bus_core"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/typ"
	"github.com/stretchr/testify/assert"
)

// -----------------------------------------
//
//	Test: Basic BusClient Functionality
//
// -----------------------------------------
func TestBusClient(t *testing.T) {
	x_init.Init()

	var receivedMessage []byte
	message := []byte("Test Message (IPv4)")

	t.Log("Creating IPv4 bus with secret key")

	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8086)
	assert.NotNil(t, bus, "Bus is nil")
	assert.NoError(t, bus.Start(), "Bus start failed")
	defer func() { assert.NoError(t, bus.Stop(), "Bus stop failed") }()

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8086)
	})(1, bus)

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			if receivedMessage != nil && string(receivedMessage) == string(message) {
				assert.Equal(t, string(message), string(receivedMessage))
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Timeout waiting for message")
	case <-done:
		t.Log("Message received successfully")
	}

	wg.Wait()
}

// -----------------------------------------
//
//	Test: Duplicate Subscription
//
// -----------------------------------------
func TestDuplicateSubscription(t *testing.T) {
	x_init.Init()

	var receivedMessage []byte

	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8085)
	assert.NotNil(t, bus, "Bus is nil")
	assert.NoError(t, bus.Start(), "Bus start failed")
	defer func() { assert.NoError(t, bus.Stop(), "Bus stop failed") }()

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8085)
	})(1, bus)

	// First handler and subscription
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	// Duplicate subscription to the same topic and queue
	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.PublishMessage("test.topic", []byte("Test Message"), "testQueue")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for receivedMessage == nil || string(receivedMessage) != "Test Message" {
			time.Sleep(10 * time.Millisecond)
		}
		assert.Equal(t, "Test Message", string(receivedMessage))
	}()
	wg.Wait()
}

// -----------------------------------------
//
//	Test: Message Processing
//
// -----------------------------------------
func TestMessageProcessing(t *testing.T) {
	x_init.Init()

	var receivedMessage []byte

	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8084)
	assert.NotNil(t, bus)
	assert.NoError(t, bus.Start(), "Bus start failed")
	defer func() { assert.NoError(t, bus.Stop(), "Bus stop failed") }()

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8084)
	})(1, bus)

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.PublishMessage("test.topic", []byte("Test Processing Message"), "testQueue")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for receivedMessage == nil || string(receivedMessage) != "Test Processing Message" {
			time.Sleep(10 * time.Millisecond)
		}
		assert.Equal(t, "Test Processing Message", string(receivedMessage))
	}()
	wg.Wait()
}

// -----------------------------------------
//
//	Test: Invalid Topic Handling
//
// -----------------------------------------
func TestInvalidTopic(t *testing.T) {
	x_init.Init()

	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8083)
	assert.NotNil(t, bus)
	assert.NoError(t, bus.Start(), "Bus start failed")
	defer func() { assert.NoError(t, bus.Stop(), "Bus stop failed") }()

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8083)
	})(1, bus)

	// Subscribing to invalid topic
	err := client.Subscribe("invalid.topic", "testQueue", func(subject string, msg []byte) {})
	assert.Error(t, err, "Expected error for invalid topic subscription")

	// Publishing to invalid topic
	err = client.PublishMessage("invalid.topic", []byte("Test Message to Invalid Topic"), "testQueue")
	assert.Error(t, err, "Expected error for publishing to invalid topic")
}
