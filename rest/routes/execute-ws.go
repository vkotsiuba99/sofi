package routes

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
	"sofi/internal/pool"
	"sofi/pkg"
)

type socketData struct {
	Language string `json:"language"`
	Content  string `json:"content"`
}

type wsResponse struct {
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
			Stdin:       []string{},
			Tests:       []pool.TestResult{},
			BypassCache: true,
		}, pipeChannel)

		for {
			select {
			case shouldTerminate := <-pipeChannel.Terminate:
				fmt.Println("terminate", shouldTerminate)
				return
			case output := <-pipeChannel.Data:
				// Send the result of the code back to the client.
				err = websocket.JSON.Send(ws, wsResponse{
					RunOutput: output,
				})
				if err != nil {
					fmt.Println("sending error:", err)
					return
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
