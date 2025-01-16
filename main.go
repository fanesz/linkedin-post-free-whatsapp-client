package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func main() {
	container, err := sqlstore.New("sqlite3", "file:examplestore.db?_foreign_keys=on", nil)
	if err != nil {
		panic(fmt.Errorf("failed to initialize sqlstore: %w", err))
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, nil)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if client.Connect() != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR code:", evt.Code)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	sendMessage(client)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}

func sendMessage(client *whatsmeow.Client) {
	targetJID := types.NewJID("xxxxxxxxxxxxx", types.DefaultUserServer)
	// replace xxxxxxxxxxxxx with the phone number of the recipient
	message := &waE2E.Message{
		Conversation: proto.String("Hello, world!"),
	}

	response, err := client.SendMessage(context.Background(), targetJID, message)
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}
	fmt.Println("Message sent successfully at:", response.Timestamp)
}
