// file: buz/bus/bus_client/bus_client_test.go

package bus_client_test

import (
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/buzz/bus/bus_client"
	"github.com/rskv-p/buzz/bus/bus_core"
	"github.com/rskv-p/buzz/bus/bus_type"
	"github.com/stretchr/testify/assert"
)

//-----------------------------------------
//  Test BusClient
//-----------------------------------------

func TestBusClient(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Log creation of the main bus with the secret key
	t.Log("Creating bus with secret key 'secret-key'")

	// Create the main bus with a secret key
	bus := bus_core.NewBus("secret-key", nil, 5)

	// Log the creation of the client through ClientFactory
	t.Log("Creating bus client with ID 1")

	// Create the client through ClientFactory
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)

	// Log registration of the handler for 'test.topic' subject
	t.Log("Registering handler for subject 'test.topic'")

	// Register the message handler for "test.topic"
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})

	// Assert the handler registration was successful
	assert.NoError(t, err, "Failed to register message handler")

	// Log subscription of the client to the topic and queue
	t.Log("Subscribing client to topic 'test.topic' with queue 'testQueue'")

	// Subscribe the client to the "test.topic" and "testQueue"
	err = client.Subscribe("test.topic", "testQueue")
	assert.NoError(t, err, "Failed to subscribe client")

	// Log publishing a message to the topic
	t.Log("Publishing message 'Test Message' to topic 'test.topic' with queue 'testQueue'")

	// Publish the message to "test.topic" with the correct queue "testQueue"
	message := []byte("Test Message")
	err = client.PublishMessage("test.topic", message, "testQueue")
	assert.NoError(t, err, "Failed to publish message")

	// Log creation of a WaitGroup to synchronize message processing
	t.Log("Creating WaitGroup to synchronize message processing")

	// Create a WaitGroup to synchronize message processing
	var wg sync.WaitGroup
	wg.Add(1) // Increment the wait counter by 1

	// Start a goroutine to wait for the message to be received
	go func() {
		defer wg.Done() // Decrement the wait counter by 1

		// Give time for asynchronous message processing
		// Wait until the message is received
		t.Log("Waiting for message to be received...")
		for receivedMessage == nil || string(receivedMessage) != "Test Message" {
			// Add a small delay to avoid busy waiting
			time.Sleep(10 * time.Millisecond)
		}

		// Log when the message is received
		t.Log("Message received successfully")

		// Assert that the message was received correctly
		assert.Equal(t, "Test Message", string(receivedMessage), "Message was not received correctly")
	}()

	// Log that the goroutine will complete after the message is received
	t.Log("Waiting for message processing to complete...")

	// Wait for the goroutine to finish
	wg.Wait()

	// Log the completion of the test
	t.Log("Test completed successfully")
}

func TestDuplicateSubscription(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Create the main bus with a secret key
	bus := bus_core.NewBus("secret-key", nil, 5)

	// Create the client through ClientFactory
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)

	// Register the message handler for "test.topic"
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register message handler")

	// Subscribe the client to the "test.topic" and "testQueue"
	err = client.Subscribe("test.topic", "testQueue")
	assert.NoError(t, err, "Failed to subscribe client to 'test.topic' with 'testQueue'")

	// Attempt to subscribe to the same topic and queue again
	err = client.Subscribe("test.topic", "testQueue")
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

	// Create the main bus with a secret key
	bus := bus_core.NewBus("secret-key", nil, 5)

	// Create the client through ClientFactory
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)

	// Register the message handler for "test.topic"
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register message handler")

	// Subscribe the client to the "test.topic" and "testQueue"
	err = client.Subscribe("test.topic", "testQueue")
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
	// Create the main bus with a secret key
	bus := bus_core.NewBus("secret-key", nil, 5)

	// Create the client through ClientFactory
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)

	// Try subscribing to a non-existing topic (invalid topic)
	err := client.Subscribe("invalid.topic", "testQueue")
	assert.Error(t, err, "Subscription to invalid topic should fail")

	// Try publishing to a non-existing topic
	message := []byte("Test Message to Invalid Topic")
	err = client.PublishMessage("invalid.topic", message, "testQueue")
	assert.Error(t, err, "Publishing to invalid topic should fail")
}
