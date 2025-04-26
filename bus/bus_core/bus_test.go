// file: buz/bus/bus_core/bus_test.go

package bus_core_test

import (
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/buzz/bus/bus_client"
	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_core"
	"github.com/rskv-p/buzz/bus/bus_type"
	"github.com/stretchr/testify/assert"
)

//-----------------------------------------
//  TestRetrySubscribe
//-----------------------------------------

// TestRetrySubscribe tests retry subscription mechanism
func TestRetrySubscribe(t *testing.T) {
	// Mock handler for the test
	var receivedMessage []byte

	// Create main bus with secret key
	bus := bus_core.NewBus("secret-key", nil, 5)
	bus_comm.Infof("Bus created with secret key")

	// Create client using ClientFactory
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)
	bus_comm.Infof("Client 1 created and ready")

	// Register message handler
	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler")
	bus_comm.Infof("Handler registered for subject 'test.topic'")

	// Subscribe client to topic
	err = client.Subscribe("test.topic", "testQueue")
	assert.NoError(t, err, "Failed to subscribe client")
	bus_comm.Infof("Client subscribed to 'test.topic' with queue 'testQueue'")

	// Retry publish message
	message := []byte("Test Message")
	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish")
	bus_comm.Infof("Message 'Test Message' published")

	// Wait for message processing using sync.WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			if receivedMessage != nil && string(receivedMessage) == "Test Message" {
				bus_comm.Infof("Message received successfully")
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
	bus_comm.Infof("TestRetrySubscribe completed")
}

//-----------------------------------------
//  TestRetryPublish
//-----------------------------------------

// TestRetryPublish tests retry publish mechanism
func TestRetryPublish(t *testing.T) {
	var receivedMessage []byte
	bus := bus_core.NewBus("secret-key", nil, 5)
	bus_comm.Infof("Bus created with secret key")

	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)
	bus_comm.Infof("Client 1 created and ready")

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler")
	bus_comm.Infof("Handler registered for subject 'test.topic'")

	err = client.Subscribe("test.topic", "testQueue")
	assert.NoError(t, err, "Failed to subscribe client")
	bus_comm.Infof("Client subscribed to 'test.topic' with queue 'testQueue'")

	message := []byte("Test Retry Message")
	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err, "Failed to retry publish")
	bus_comm.Infof("Message 'Test Retry Message' published")

	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			bus_comm.Infof("Checking for received message...")
			if receivedMessage != nil && string(receivedMessage) == "Test Retry Message" {
				bus_comm.Infof("Message received successfully")
				assert.Equal(t, "Test Retry Message", string(receivedMessage), "Message mismatch")
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		bus_comm.Errorf("Timeout while waiting for message")
		t.Fatal("Timeout while waiting for message")
	case <-done:
		t.Log("Test completed successfully")
	}

	wg.Wait()
	bus_comm.Infof("TestRetryPublish completed")
}

//-----------------------------------------
//  TestSubscriptionErrorHandling
//-----------------------------------------

// TestSubscriptionErrorHandling tests error handling when subscribing
func TestSubscriptionErrorHandling(t *testing.T) {
	var ReceivedMessage []byte
	bus := bus_core.NewBus("secret-key", nil, 5)
	bus_comm.Infof("Bus created with secret key")
	print(ReceivedMessage)
	client := bus_type.ClientFactory(func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5)
	})(1, bus)
	bus_comm.Infof("Client created and ready")

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		ReceivedMessage = msg
	})
	assert.NoError(t, err, "Failed to register handler")
	bus_comm.Infof("Handler registered for 'test.topic'")

	err = client.Subscribe("test.topic", "testQueue")
	assert.NoError(t, err, "Failed to subscribe")
	bus_comm.Infof("Client subscribed to 'test.topic' with queue 'testQueue'")

	err = client.Subscribe("test.topic", "testQueue")
	assert.NoError(t, err, "Failed to subscribe again")
	bus_comm.Infof("Client already subscribed to 'test.topic'")

	// Try subscribing to invalid topic
	bus_comm.Infof("Subscribing to invalid topic 'invalid.topic'")
	err = client.Subscribe("invalid.topic", "testQueue")
	if err == nil {
		bus_comm.Errorf("Expected error, got nil")
		t.Fatal("Expected error, got nil")
	}
	bus_comm.Errorf("Error subscribing to invalid topic: %v", err)

	// Try publishing to invalid topic
	message := []byte("Test Message")
	bus_comm.Infof("Publishing message to invalid topic")
	err = client.PublishMessage("invalid.topic", message, "testQueue")
	if err == nil {
		bus_comm.Errorf("Expected error, got nil")
		t.Fatal("Expected error, got nil")
	}
	bus_comm.Errorf("Error publishing to invalid topic: %v", err)

	bus_comm.Infof("Skipping message reception due to invalid publish")

	bus_comm.Infof("TestSubscriptionErrorHandling completed")
}
