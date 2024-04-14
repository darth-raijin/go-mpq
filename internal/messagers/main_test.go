package messagers_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var rabbitMQConnectionString string

func TestMain(m *testing.M) {
	ctx := context.Background()
	rabbitMQContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3.11-management",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("started TCP listener on [::]:5672").WithStartupTimeout(120 * time.Second), // Adjust the timeout as necessary
		},
		Started: true,
	})
	if err != nil {
		panic(err) // Properly handle initialization errors
	}
	defer rabbitMQContainer.Terminate(ctx) // Ensure clean-up

	// Create the 'test' vhost within the RabbitMQ container
	exitCode, outputReader, execErr := rabbitMQContainer.Exec(ctx, []string{"rabbitmqctl", "add_vhost", "test"})
	if execErr != nil || exitCode != 0 {
		panic(fmt.Sprintf("Failed to create vhost: %v, exit code: %d", execErr, exitCode))
	}
	readOutput(outputReader)

	// Set permissions for the 'guest' user on the 'test' vhost
	exitCode, outputReader, execErr = rabbitMQContainer.Exec(ctx, []string{"rabbitmqctl", "set_permissions", "-p", "test", "guest", ".*", ".*", ".*"})
	if execErr != nil || exitCode != 0 {
		panic(fmt.Sprintf("Failed to set permissions: %v, exit code: %d", execErr, exitCode))
	}
	readOutput(outputReader)

	// Retrieve host and port to construct the RabbitMQ connection string
	host, err := rabbitMQContainer.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := rabbitMQContainer.MappedPort(ctx, "5672")
	if err != nil {
		panic(err)
	}
	rabbitMQConnectionString = fmt.Sprintf("amqp://guest:guest@%s:%s/test", host, port.Port())

	// Execute the tests
	code := m.Run()
	os.Exit(code)
}

func readOutput(r io.Reader) {
	if r != nil {
		output, err := io.ReadAll(r)
		if err != nil {
			fmt.Printf("Error reading output: %v\n", err)
		} else {
			fmt.Printf("Command output: %s\n", output)
		}
	}
}
