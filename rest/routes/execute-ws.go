package routes

import (
	"fmt"
	"github.com/labstack/echo/v4"
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
				// Send the result of the code back to the client.
				err = websocket.JSON.Send(ws, wsResponse{
					Type:      "output",
					RunOutput: output,
				})
				if err != nil {
					fmt.Println("sending error:", err)
					return
				}
			case <-pipeChannel.Terminate:
				err = websocket.JSON.Send(ws, wsResponse{
					Type:      "terminate",
					RunOutput: "",
				})
				if err != nil {
					fmt.Println("sending error:", err)
					return
				}
				return
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
