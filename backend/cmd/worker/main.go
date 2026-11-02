// Command worker consume mensajes de Azure Service Bus.
// No forma parte de la demo del TFG. La cadena de conexión se lee de
// SERVICEBUS_CONNECTION_STRING y no se guarda en el código.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

const queueName = "audio-jobs"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	serviceBusConn := os.Getenv("SERVICEBUS_CONNECTION_STRING")
	if serviceBusConn == "" {
		slog.Error("missing_servicebus_connection")
		os.Exit(1)
	}

	client, err := azservicebus.NewClientFromConnectionString(serviceBusConn, nil)
	if err != nil {
		slog.Error("servicebus_client_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close(context.Background())

	receiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		slog.Error("servicebus_receiver_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer receiver.Close(context.Background())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("worker_started", slog.String("queue", queueName))

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker_stopping")
			return
		default:
		}

		messages, err := receiver.ReceiveMessages(ctx, 1, nil)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("receive_failed", slog.String("error", err.Error()))
			time.Sleep(2 * time.Second)
			continue
		}

		for _, msg := range messages {
			body := string(msg.Body)
			slog.Info("processing_job", slog.String("body", body))
			// Simula trabajo (transcripción / feedback en un escenario real).
			time.Sleep(2 * time.Second)
			if err := receiver.CompleteMessage(ctx, msg, nil); err != nil {
				slog.Error("complete_failed", slog.String("error", err.Error()))
				continue
			}
			slog.Info("job_completed", slog.String("body", body))
		}
	}
}
