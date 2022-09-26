package sofi

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sofi/sandbox"
	"syscall"
	"time"
)

func main() {
	app := &cli.App{
		Name:  "sofi",
		Usage: "Use the sofi code execution engine to run your code.",
		Commands: []*cli.Command{
			{
				Name:  "execute",
				Usage: "Execute a test sofi code",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l", "lang"},
						Value:   "python",
						Usage:   "set the language for the sofi sandbox runner.",
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Value:       "",
						DefaultText: "execute an example code.",
						Usage:       "set the specific file that should be executed.",
					},
				},
				Action: func(ctx *cli.Context) error {
					language := ctx.String("language")

					var runner *sandbox.Runner
					for _, r := range sandbox.Runners {
						if language == r.Name {
							runner = &r
							break
						}
					}

					if runner == nil {
						return fmt.Errorf("no language found with name %s", language)
					}

					filePath := ctx.String("file")
					code, err := extractCodeOfFile(filePath)
					if err != nil {
						return fmt.Errorf("something went wrong while reading the file %s", filePath)
					}

					execute(runner, code)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func extractCodeOfFile(filePath string) (string, error) {
	if filePath == "" {
		return "", nil
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	code := string(content)
	return code, nil
}

func execute(runner *sandbox.Runner, code string) {
	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			default:
				fmt.Println("signal received", s)
			}
		}
	}()

	if code == "" {
		code = runner.ExampleCode
	}

	s, err := sandbox.NewSandbox(runner.Name, []byte(code))
	if err != nil {
		panic(err)
	}
	defer s.Clean()

	stopTicking := make(chan bool)
	go func() {
		for t := range time.Tick(time.Second * 1) {
			select {
			case <-stopTicking:
				return
			default:
				fmt.Println("ticking", t)
				h, _ := time.ParseDuration("5s")
				expireTime := s.LastTimestamp.Add(h)
				if expireTime.Before(time.Now()) {
					s.Clean()
				}
			}
		}
	}()

	output, err := s.Run()
	if err != nil {
		panic(err)
	}

	for _, op := range output {
		fmt.Println(op)
	}

	stopTicking <- true
}
