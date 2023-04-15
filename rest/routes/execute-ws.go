package routes

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"sofi/internal/pool"
	"sofi/pkg"
)

type socketData struct {
	Language string   `json:"language" binding:"required"`
	Content  string   `json:"content" binding:"required"`
	Stdin    []string `json:"stdin,omitempty"`
}

type wsResponse struct {
	Type      string `json:"type"`
	RunOutput string `json:"runOutput"`
}

func ExecuteWs(c echo.Context, rceEngine *pkg.RceEngine) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// Receive and parse send JSON data from the client.
		data := socketData{}
		err := websocket.JSON.Receive(ws, &data)
		if err != nil {
			fmt.Println("receiving error:", err)
			return
		}

		// Execute the code of the client.
		pipeChannel := pkg.PipeChannel{
			Data:      make(chan string),
			Terminate: make(chan bool),
		}
		go rceEngine.DispatchStream(pool.WorkData{
			Lang:        data.Language,
			Code:        data.Content,
			Stdin:       data.Stdin,
			Tests:       []pool.TestResult{},
			BypassCache: true,
		}, pipeChannel)

		for {
			select {
			case output := <-pipeChannel.Data:
				response := wsResponse{
					Type:      "output",
					RunOutput: output,
				}

				// Send the result of the code back to the client.
				err = websocket.JSON.Send(ws, response)
				if err != nil {
					fmt.Println("sending error:", err)
					return
				}

				logResponse(data, response)
			case <-pipeChannel.Terminate:
				response := wsResponse{
					Type:      "terminate",
					RunOutput: "",
				}
				err = websocket.JSON.Send(ws, response)
				if err != nil {
					fmt.Println("sending error:", err)
					return
				}

				logResponse(data, response)
				return
			}
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

// logResponse takes in the socket data as a request and the to be logged response for
// that request. It will log to the specific logger with `pkg.Logger`.
func logResponse(request socketData, response wsResponse) {
	dataBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Println("marshalling error:", err)
		return
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("marshalling error:", err)
		return
	}

	pkg.Logger.Info(
		"ws-request",
		zap.String("requestBody", string(dataBytes)),
		zap.String("responseBody", string(responseBytes)),
	)
}
