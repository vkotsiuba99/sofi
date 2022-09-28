# sofi

sofi is a remote docker based sandbox code execution engine written in Go.

Currently it supports the following languages:

- Go (golang:1.18-alpine)
- Java (openjdk:8u232-jdk)
- C (gcc:latest)
- C++ (gcc:latest)
- Python (python:3.9.1-alpine)
- JavaScript (node:lts-alpine)
- JavaScript (node:lts-alpine & latest tsc)
- Julia (julia:1.7.1-alpine)

## Installation

For the installation of sofi, you need to have Docker and Go installed.

If you want to have the latest sofi image on your machine, execute the [`build_sofi_image.sh`](https://github.com/vkotsiuba99/sofi/tree/master/build/build_sofi_image.sh) script.

In addition, you need to pull the latest images by executing [`pull_images.sh`](https://github.com/vkotsiuba99/sofi/tree/master/build/pull_images.sh). This will pull all the docker images that are being used by sofi. This step should only be executed once.

## Usage

You can feel free to run the CLI by executing the `main.go` file with the following command:

```sh
$ go run main.go
```

This will prompt you with some flags and commands you can use.

### Commands and Flags

The following section contains all the commands and flags that can be used while running the CLI.

<details>
  <summary>Execute</summary>

  <p>
    The execute command will execute code in a containerized sandbox.
  </p>

  | Flag | Aliases | Description                                    | Default |
  |---|------------------------------------------------|---|---|
  | --language | -l, -lang | Set the language for the sofi sandbox runner.  | python |
  | --file | -f | Set the specific file that should be executed. | example code in runner struct |
</details>

## License

[MIT](https://choosealicense.com/licenses/mit/)