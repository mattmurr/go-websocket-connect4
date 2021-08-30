package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWsServer(t *testing.T) {
	t.Run("Can connect to WebSocket server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		url := "ws" + strings.TrimPrefix(server.URL, "http") + "/"

		ws, res, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Errorf("Could not open a ws connection on '%s' '%v'", url, err)
		}
		const expectedStatus = http.StatusSwitchingProtocols
		if res.StatusCode != expectedStatus {
			t.Errorf("Expected status '%d', got '%d'", expectedStatus, res.StatusCode)
		}
		defer ws.Close()

		t.Run("Subscribes callback to 'test' event successfully", func(t *testing.T) {
			randInt := rand.Intn(10)

			testCallback := func(c *websocket.Conn, messageType int, msg string) error {
				if len(msg) != 0 {
					t.Errorf("Expected message to be '%s', got '%s'", "", msg)
				}
				return c.WriteMessage(messageType, []byte(fmt.Sprintf("%d", randInt)))
			}

			const testEvent = "test"

			subscribe(testEvent, testCallback)
			if val, ok := subscriptions[testEvent]; !ok || len(val) != 1 {
				t.Errorf("Failed to subscribe to 'test' event")
			}

			t.Run("Sends message successfully", func(t *testing.T) {
				const expectedMessageType = websocket.TextMessage
				message := []byte("{\"event\":\"test\",\"data\":\"\"}")
				expectedMessage := []byte(fmt.Sprintf("%d", randInt))

				err = ws.WriteMessage(expectedMessageType, message)
				if err != nil {
					t.Errorf("Unable to write message '%v'", err)
				}

				t.Run("Recieves the same message and message type in response", func(t *testing.T) {
					mt, message, err := ws.ReadMessage()
					if err != nil {
						t.Errorf("Could not read message '%v'", err)
					} else {
						if mt != expectedMessageType {
							t.Errorf("Expected message type '%d', got '%d'", expectedMessageType, mt)
						}
						if bytes.Compare(message, expectedMessage) != 0 {
							t.Errorf("Expected message '%s', got '%s'", expectedMessage, message)
						}
					}
				})
			})

			t.Run("Subscribes a second callback to 'test' event successfully", func(t *testing.T) {
				subscribe(testEvent, testCallback)
				if val, found := subscriptions[testEvent]; !found || len(val) != 2 {
					t.Errorf("Failed to subscribe to 'test' event")
				}

			})
		})
	})
}
