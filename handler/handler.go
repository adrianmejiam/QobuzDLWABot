package handler

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

type Handler struct {
	Client         *whatsmeow.Client
	eventHandlerID uint32
	baseUrl        string
	qobuzPath	   string
	prefix 		   string
}

func (handler *Handler) register() {
	handler.eventHandlerID = handler.Client.AddEventHandler(handler.EventHandler)
}

func (handler *Handler) Initialize() {
	handler.prefix = os.Getenv("PREFIX")
	handler.baseUrl = os.Getenv("BASE_URL")
	handler.qobuzPath = os.Getenv("QOBUZ_PATH")
}

func (handler *Handler) EventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		handler.HandleMessage(v)
		break
	}
}

func (handler *Handler) HandleMessage(e *events.Message) {
	message := e.Message.GetConversation()

	if e.Info.IsFromMe || len(message) == 0 || string(message[0]) != handler.prefix {
		return
	}
	cmd := strings.Replace(strings.Split(message, " ")[0], handler.prefix, "", 1)
	args := strings.Split(message, " ")[1:]

	switch cmd {
	case "dl", "download":
		handler.CommandDownload(e, args)
		return
	case "s", "search":
		handler.CommandSearch(e)
		return
	}

	handler.SendMessage(e, "Command not found.")
}

func (handler *Handler) SendMessage(e *events.Message, message string) {
	_, err := handler.Client.SendMessage(context.Background(), e.Info.Chat,
		&waE2E.Message{
			Conversation: proto.String(message),
			//ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			//	Text: proto.String(message),
			//	ContextInfo: &waE2E.ContextInfo{
			//		StanzaID:    proto.String(e.Info.ID),
			//		Participant: proto.String(e.Info.Sender.ToNonAD().String()),
			//	},
			//},
		},
	)
	if err != nil {
		panic(err)
	}
}

func (handler *Handler) GetTextMessage(filename string) (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	file, err := os.Open(path.Join(workdir, "messages", filename))
	if err != nil {
		return "", err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := io.ReadAll(file)
	return string(b), nil
}
