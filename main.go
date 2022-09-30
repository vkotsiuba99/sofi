package sofi

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sofi/file"
	"sofi/sandbox"
	"strconv"
	"strings"
	"sync"
	"time"
)

func executeCode(wg *sync.WaitGroup) {
	var sandboxLang *sandbox.Language
	for _, l := range sandbox.Languages {
		if "python" == l.Name {
			sandboxLang = &l
			break
		}
	}

	mainCode, err := file.ExtractCodeOfFile("examples/python/app.py")
	s, output, err := sandbox.Run(sandboxLang, mainCode, make([]sandbox.SandboxFile, 0), make([]sandbox.SandboxFile, 0))
	if err != nil {
		fmt.Printf("something went wrong while executing sandbox runner %s", err)
		return
	}

	defer wg.Done()
	defer s.Clean()

	fmt.Println("\n=== BUILD OUTPUT ===")
	fmt.Printf("Error: %s, Body: %s\n\n", strconv.FormatBool(output.BuildError), output.BuildBody)
	fmt.Println("=== RUN OUTPUT ===")
	fmt.Printf("Error: %s, Body: %s\n", strconv.FormatBool(output.RunError), output.RunBody)
	fmt.Println("=== TEST OUTPUT ===")
	fmt.Printf("Error: %s, Body: \n%s\n", strconv.FormatBool(output.TestError), output.TestBody)
}

func benchmark() {
	max := 2
	wg := new(sync.WaitGroup)
	wg.Add(max)
	for i := 0; i < max; i++ {
		go executeCode(wg)
	}

	wg.Wait()
}

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
						Name:        "main",
						Aliases:     []string{"m"},
						Value:       "",
						DefaultText: "default main file from example code.",
						Usage:       "set the specific main file for executing first.",
					},
					&cli.StringFlag{
						Name:        "dir",
						Aliases:     []string{"d"},
						Value:       "",
						DefaultText: "copies all the example code.",
						Usage:       "set the specific directory that should be copied and executed from.",
					},
				},
				Action: func(ctx *cli.Context) error {
					language := strings.ToLower(ctx.String("language"))

					var sandboxLang *sandbox.Language
					for _, l := range sandbox.Languages {
						if language == l.Name {
							sandboxLang = &l
							break
						}
					}

					if sandboxLang == nil {
						return fmt.Errorf("no language found with name %s", language)
					}

					mainFilePath := ctx.String("main")
					mainCode, err := file.ExtractCodeOfFile(mainFilePath)
					if err != nil {
						return fmt.Errorf("something went wrong while reading the main file path %s", mainFilePath)
					}

					dirPath := ctx.String("dir")
					sandboxTests := make([]sandbox.SandboxFile, 0)
					files := make([]sandbox.SandboxFile, 0)

					if dirPath != "" {
						dirFiles, err := os.ReadDir(dirPath)
						if err != nil {
							return fmt.Errorf("something went wrong while reading the dir path %s", dirPath)
						}

						for _, dirFile := range dirFiles {
							if dirPath+dirFile.Name() == mainFilePath {
								continue
							}

							fileCode, err := file.ExtractCodeOfFile(dirPath + dirFile.Name())
							if err != nil {
								return fmt.Errorf("something went wrong while reading the file %s", dirPath+dirFile.Name())
							}

							if strings.Contains(strings.ToLower(dirFile.Name()), "test") {
								sandboxTests = append(sandboxTests, sandbox.SandboxFile{Code: []byte(fileCode), FileName: dirFile.Name()})
							} else {
								files = append(files, sandbox.SandboxFile{Code: []byte(fileCode), FileName: dirFile.Name()})
							}
						}
					}

					start := time.Now()
					s, output, err := sandbox.Run(sandboxLang, mainCode, files, sandboxTests)
					if err != nil {
						return fmt.Errorf("something went wrong while executing sandbox runner %s", err)
					}
					defer s.Clean()

					fmt.Println("\n=== BUILD OUTPUT ===")
					fmt.Printf("Error: %s, Body: %s\n\n", strconv.FormatBool(output.BuildError), output.BuildBody)
					fmt.Println("=== RUN OUTPUT ===")
					fmt.Printf("Error: %s, Body: %s\n", strconv.FormatBool(output.RunError), output.RunBody)
					fmt.Println("=== TEST OUTPUT ===")
					fmt.Printf("Error: %s, Body: \n%s\n", strconv.FormatBool(output.TestError), output.TestBody)

					duration := time.Since(start)
					fmt.Printf("duration before clean=%s\n", duration)

					s.Clean()

					duration = time.Since(start)
					fmt.Printf("duration after clean=%s\n", duration)

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
