// file: buzz/mod/m_bus/bus_core/bus_core_test.go

package bus_core_test

import (
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/buzz/mod/m_bus/bus_client"
	"github.com/rskv-p/buzz/mod/m_bus/bus_core"
	"github.com/rskv-p/buzz/mod/m_bus/bus_sub"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
	"github.com/rskv-p/subtree"
	"github.com/stretchr/testify/assert"
)

//-----------------------------------------
//  SubjectTree Initialization
//-----------------------------------------

func TestNewSubjectTree(t *testing.T) {
	tree := subtree.NewSubjectTree[*typ.Subscription]()
	if tree == nil {
		t.Fatal("Failed to initialize SubjectTree")
	}
}

//-----------------------------------------
//  Sublist Initialization
//-----------------------------------------

func TestNewSublist(t *testing.T) {
	sublist := bus_sub.NewSublist(64)
	if sublist.Tree == nil {
		t.Fatal("Tree should not be nil")
	}
	if sublist.ExactCache == nil {
		t.Fatal("ExactCache should not be nil")
	}
	if sublist.WildcardCache == nil {
		t.Fatal("WildcardCache should not be nil")
	}
}

//-----------------------------------------
//  RetrySubscribe Test (IPv4)
//-----------------------------------------

func TestRetrySubscribe(t *testing.T) {
	x_init.Init()

	var receivedMessage []byte

	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8087)
	assert.NoError(t, bus.Start(), "Bus start failed (IPv4)")
	defer func() { assert.NoError(t, bus.Stop(), "Bus stop failed (IPv4)") }()
	x_log.Info("Bus created (IPv4)")

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8087)
	})(1, bus)

	assert.NotNil(t, client, "Client is nil (IPv4)")

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	message := []byte("Test Message (IPv4)")
	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			if receivedMessage != nil && string(receivedMessage) == "Test Message (IPv4)" {
				assert.Equal(t, "Test Message (IPv4)", string(receivedMessage))
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Timeout waiting for message (IPv4)")
	case <-done:
		t.Log("Message received successfully (IPv4)")
	}

	wg.Wait()
}

//-----------------------------------------
//  RetryPublish Test (IPv4)
//-----------------------------------------

func TestRetryPublish(t *testing.T) {
	x_init.Init()

	var receivedMessage []byte

	bus := bus_core.NewBus("secret-key", 5, "127.0.0.1", 8088)
	assert.NotNil(t, bus, "Bus is nil (IPv4)")
	assert.NoError(t, bus.Start(), "Bus start failed (IPv4)")
	defer func() { assert.NoError(t, bus.Stop(), "Bus stop failed (IPv4)") }()
	x_log.Info("Bus created (IPv4)")

	client := typ.ClientFactory(func(id uint64, bus typ.IBus) typ.IBusClient {
		return bus_client.NewBusClient(id, bus, "client-secret-key", 5, "127.0.0.1", 8088)
	})(1, bus)

	assert.NotNil(t, client, "Client is nil (IPv4)")

	err := client.RegisterHandler("test.topic", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	err = client.Subscribe("test.topic", "testQueue", func(subject string, msg []byte) {
		receivedMessage = msg
	})
	assert.NoError(t, err)

	message := []byte("Test Retry Message")
	err = client.RetryPublish("test.topic", message, 3, 500*time.Millisecond, "testQueue")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	done := make(chan struct{})
	timeout := time.After(10 * time.Second)

	go func() {
		defer wg.Done()
		for {
			if receivedMessage != nil && string(receivedMessage) == "Test Retry Message" {
				assert.Equal(t, "Test Retry Message", string(receivedMessage))
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Timeout waiting for message (IPv4)")
	case <-done:
		t.Log("Message received successfully (IPv4)")
	}

	wg.Wait()
}
