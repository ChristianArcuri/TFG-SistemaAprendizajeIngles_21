// Command enqueue envía mensajes a la cola audio-jobs.
// No forma parte de la demo del TFG. La cadena de conexión se lee de
// SERVICEBUS_CONNECTION_STRING y no se guarda en el código.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

const (
	queueName    = "audio-jobs"
	messageCount = 50
)

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

	sender, err := client.NewSender(queueName, nil)
	if err != nil {
		slog.Error("servicebus_sender_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer sender.Close(context.Background())

	ctx := context.Background()
	for i := 0; i < messageCount; i++ {
		msg := &azservicebus.Message{
			Body: []byte(fmt.Sprintf(
				`{"job":"audio-analyze","id":%d,"at":"%s"}`,
				i,
				time.Now().UTC().Format(time.RFC3339),
			)),
		}
		if err := sender.SendMessage(ctx, msg, nil); err != nil {
			slog.Error("send_failed", slog.Int("id", i), slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	slog.Info("messages_sent", slog.String("queue", queueName), slog.Int("count", messageCount))
}
