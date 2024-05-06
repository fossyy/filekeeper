package uploadHandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	filesView "github.com/fossyy/filekeeper/view/upload"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

var log *logger.AggregatedLogger
var mu sync.Mutex

func init() {
	log = logger.Logger()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSAuth struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type WSData struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Index int    `json:"index"`
	Chunk string `json:"chunk"`
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := filesView.Main("upload page")
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func WEBSOCKET(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading from websocket:", err)
		return
	}
	fmt.Println(string(p))

	var Auth WSAuth
	err = json.Unmarshal(p, &Auth)
	if err != nil {
		fmt.Println("Error unmarshalling json:", err)
		return
	}
	switch Auth.Type {
	case "auth":
		status, _, _ := session.GetSessionWithID(Auth.Token)
		switch status {
		case session.Authorized:
			conn.WriteMessage(messageType, []byte("Authorized"))
			HandleConnection(conn, r)
		case session.Unauthorized:
			conn.WriteMessage(messageType, []byte("Unauthorized"))
			return
		default:

		}
	default:
		return
	}

	return
}

func HandleConnection(conn *websocket.Conn, r *http.Request) {
	total := 0
	id := r.PathValue("id")
	info, err := cache.GetFile(id)
	if err != nil {
		fmt.Println("Error geting the file ", err)
		return
	}

	defer conn.Close()
	dst, err := os.OpenFile("./"+info.Name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("error opening the file :", err)
		return
	}
	defer dst.Close()

	for {
		_, infoUpload, _ := conn.ReadMessage()
		fmt.Println(string(infoUpload))
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("Connection closed by user:", err)
				return
			}
			fmt.Println("Error reading message:", err)
			return
		}
		total += len(p)
		io.Copy(dst, bytes.NewReader(p))
		fmt.Println(total)
		if total >= info.Size {
			return
		}
		if err := conn.WriteMessage(1, []byte(time.Now().String())); err != nil {
			fmt.Println("Error writing message:", err)
			return
		}
	}
}
