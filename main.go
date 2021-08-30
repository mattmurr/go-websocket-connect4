package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type subscriptionCallback func(*websocket.Conn, int, string) error

var subscriptions = make(map[string]([]subscriptionCallback))

func subscribe(event string, callback subscriptionCallback) {
	if arr, found := subscriptions[event]; !found {
		newArr := make([]subscriptionCallback, 1)
		newArr[0] = callback
		subscriptions[event] = newArr
	} else {
		subscriptions[event] = append(arr, callback)
	}
}

var upgrader = websocket.Upgrader{}

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error '%v'", err)
	}
	defer c.Close()

	for {
		mt, msgRaw, err := c.ReadMessage()
		if err != nil {
			log.Printf("Read error '%v'", err)
			break
		}

		msg := struct {
			Event string `json:"event"`
			Data  string `json:"data"`
		}{}

		err = json.Unmarshal(msgRaw, &msg)
		if err != nil {
			log.Printf("Unmarshall error '%v'", err)
			break
		}

		if arr, found := subscriptions[msg.Event]; found {
			for i := 0; i < len(arr); i++ {
				log.Println(arr)
				arr[i](c, mt, msg.Data)
			}
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Print("🚀 Listening on: http://127.0.0.1:3000 🚀")
	log.Fatal(http.ListenAndServe("127.0.0.1:3000", nil))
}
