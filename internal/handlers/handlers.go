package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	jet "github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayloadDTO)

var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		slog.Error(err.Error())
	}
}

type WebSocketConnection struct {
	*websocket.Conn
}

type WsUpgradeResponseDTO struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayloadDTO struct {
	Username string              `json:"username"`
	Action   string              `json:"action"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

func WsUpgrade(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Client connected to endpoint")

	var response WsUpgradeResponseDTO
	response.Message = `<em><small> Connected to server </small><em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		slog.Error(err.Error())
	}

	go ListenForWs(&conn)
}

func ListenToWsChannel() {
	var response WsUpgradeResponseDTO

	for {
		event := <-wsChan
		switch event.Action {
		case "username":
			clients[event.Conn] = event.Username
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcast(response)
		case "left":
			response.Action = "list_users"
			delete(clients, event.Conn)
			response.ConnectedUsers = getUserList()
			broadcast(response)
		case "message":
			response.Action = "message"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", event.Username, event.Message)
			broadcast(response)
		}

	}
}

func getUserList() []string {
	var usersList []string
	for _, username := range clients {
		if username != "" {
			usersList = append(usersList, username)
		}
	}

	sort.Strings(usersList)
	return usersList
}

func broadcast(response WsUpgradeResponseDTO) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			slog.Error(fmt.Sprintf("Cannot broad cast message: %s", err.Error()))
			_ = client.Close()
			delete(clients, client)

		}
	}
}

func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error(fmt.Sprintf("Error:%v", r))
		}
	}()

	var payload WsPayloadDTO

	for {
		err := conn.ReadJSON(&payload)
		if err == nil {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}
