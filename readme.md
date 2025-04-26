# Buz - a Scalable and Reliable Message Bus System 🔹

![Buzz Lightyear](.bus/md/buzz.jpeg)


## Overview 🌐
**Buzz** is a highly performant, scalable, and reliable message bus system designed for modern distributed applications. It facilitates seamless communication between microservices, ensuring that messages are reliably delivered and processed in a fast and fault-tolerant manner. 

With Buzz, you can efficiently handle the message traffic of complex systems while maintaining high availability and low latency.

**Note**: Buzz is currently in the **early stages of development** and is actively being improved. It is not yet feature-complete, but contributions are welcome!

## Key Features 🚀

### 1. **Efficient Message Transport** 📡
Buzz provides reliable message delivery between distributed services. It ensures that messages are delivered even in the face of network interruptions or failures, ensuring high availability and fault tolerance for your system.

- **Real-time Messaging:** Supports real-time message passing.
- **Guaranteed Delivery:** Built-in mechanisms to ensure message reliability.

### 2. **Configurable Logging** 📑
With Buzz, you have complete control over the logging system. The application comes with built-in logging functionality that tracks and logs every message, event, and error.

- **Customizable Log Levels:** Control log verbosity (e.g., info, debug, error).
- **Log Rotation:** Automatically rotate logs based on size and age to prevent disk overuse.
- **Log Formats:** Choose between JSON or plain text for log output.

### 3. **Log Rotation and Management** 🔄
Logs can quickly grow in size, especially in production environments. Buzz handles this by providing log rotation functionality, which includes:

- **Max File Size:** Limit the size of each log file.
- **Retention Policies:** Keep only the necessary number of old log files.
- **Compression:** Compress old log files to save storage space.

### 4. **Fault Tolerance** 🛠️
Buzz is designed to be highly resilient, even in the case of service or network failures. The system ensures that no messages are lost and that services can continue processing after an unexpected failure.

- **Automatic Recovery:** Buzz can recover from failures without losing messages.
- **Retry Mechanisms:** Failed operations are retried automatically.

### 5. **Middleware Support** 🔧
Buzz allows you to extend its functionality with middleware, which can be used for tasks like authentication, permissions, logging, and more.

- **Custom Middleware:** Easily implement custom middleware to meet specific application requirements.
- **Built-in Authentication:** Secure communication with middleware for user and service authentication.

### 6. **Scalable and Extensible Architecture** ⚙️
Buzz is built to scale. Whether you're running a single instance or deploying it across multiple services in a cloud environment, Buzz can handle the load.

- **Horizontal Scaling:** Add more instances of the bus to handle increased load.
- **Extensible Framework:** Easily add new features and adapt to changing system requirements.

### 7. **Easy Integration with Microservices** 🌍
Buzz seamlessly integrates with microservices architecture, allowing different components of your system to communicate effortlessly.

- **Publish/Subscribe Model:** Components can subscribe to specific topics and publish messages.
- **Event-Driven:** Perfect for event-driven systems that need to react to changes or triggers in real-time.

### 8. **Performance Optimized** ⚡
Buzz is optimized for speed and low latency, making it ideal for high-performance applications that require real-time messaging.

- **Low Latency Messaging:** Designed to minimize the time taken for message delivery.
- **High Throughput:** Capable of handling a large number of messages per second.

## How It Works 💡

Buzz follows a publish/subscribe model. Services can publish messages to specific topics, and other services can subscribe to those topics to receive messages. The system is designed to ensure that messages are routed to the correct services and processed efficiently.

### Example Workflow:
1. **Publisher** sends a message to a specific **topic**.
2. **Subscribers** to that topic receive the message and process it.
3. The message is acknowledged once processed, and in case of failure, Buzz ensures automatic retries and fault recovery.

## Example Code 💻

Here’s a simple example of how to use Buzz for publishing and subscribing to messages:

```go
package main

import (
    "fmt"
    "github.com/yourusername/buzz/bus_comm"
)

func main() {
    // Initialize the Buzz message bus system
    bus, err := bus_comm.NewBus("secret-key", nil, 5)
    if err != nil {
        fmt.Printf("Failed to initialize bus: %v\n", err)
        return
    }

    // Start processing messages
    bus.ProcessMessages()

    // Example: Publish a message
    message := []byte("Hello, Buzz!")
    err = bus.Publish([]byte("test.topic"), []byte("testQueue"), message)
    if err != nil {
        fmt.Printf("Failed to publish message: %v\n", err)
    }
}
```

## Contributing 🤝

We welcome contributions to Buzz! The project is still in its **early stages of development**, and we are actively working on adding new features and improving performance. To contribute, please fork the repository, create a branch, and submit a pull request. All contributions will be reviewed before being merged into the main branch.

### License 📄

Buzz is open-source and licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

