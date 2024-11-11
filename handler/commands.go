package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"
	"github.com/carlmjohnson/requests"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

func (handler *Handler) CommandSearch(e *events.Message) {
	handler.SendMessage(e, fmt.Sprintf("You can find Qobuz music at the link below and use the %sdl command\nhttps://www.qobuz.com/us-en/shop",
		handler.prefix))
}

func (handler *Handler) CommandDownload(e *events.Message, args []string) {
	handler.SendMessage(e, "Please wait...")
	
	var dirname string
	err := requests.
		URL(fmt.Sprintf("%sdownload?url=%s", handler.baseUrl, strings.Join(args, " "))).
		ToString(&dirname).
		Fetch(context.Background())

	if err != nil {
		panic(err)
	}
	
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	
	dir := fmt.Sprintf("%s/%s/%s", wd, handler.qobuzPath, dirname)
	files, err := os.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
    	if file.IsDir() {
	    	continue
	    }

        data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, file.Name()))
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		uploaded, err := handler.Client.Upload(context.Background(), data, whatsmeow.MediaDocument)
		if err != nil {
			log.Fatalf("Failed to upload file: %v", err)
		}
		msg := &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
			 	Caption:       proto.String(""), 
				FileName:      proto.String(file.Name()),
			 	Mimetype:      proto.String(http.DetectContentType(data)),
			 	FileLength:    proto.Uint64(uint64(len(data))),
					
			 	DirectPath:    proto.String(uploaded.DirectPath),
			 	URL:           proto.String(uploaded.URL), 
			 	MediaKey:      uploaded.MediaKey,
				FileSHA256:	   uploaded.FileSHA256,
				FileEncSHA256: uploaded.FileEncSHA256,
			},
		}
		_, err = handler.Client.SendMessage(context.Background(), e.Info.Chat, msg)
		if err != nil {
			log.Fatalf("Failed to send file: %v", err)
		}
    }
}
