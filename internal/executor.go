package internal

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	maxOutputBufferCapacity = "65332"
)

var user = 1

type CodeOutput struct {
	User        string
	TempDirName string
	Result      string
	Error       string
}

func RunCode(lang, code string) (CodeOutput, error) {
	language, err := GetLanguageByName(lang)
	if err != nil {
		return CodeOutput{}, err
	}

	currentUser := fmt.Sprintf("user%d", user)
	tempDirName := uuid.New().String()

	err = CreateTempDir(currentUser, tempDirName)
	if err != nil {
		return CodeOutput{}, err
	}

	filename, err := CreateTempFile(currentUser, tempDirName, language.Extension)
	if err != nil {
		return CodeOutput{}, err
	}

	err = WriteToFile(filename, code)
	if err != nil {
		return CodeOutput{}, err
	}

	output, errorString := executeFile(currentUser, filename, language)

	return CodeOutput{
		User:        currentUser,
		TempDirName: tempDirName,
		Result:      output,
		Error:       errorString,
	}, nil
}

func CleanUp(currentUser, tempDirName string) {
	DeleteTempDir(currentUser, tempDirName)
	cleanProcesses(currentUser)
	restoreUserDir(currentUser)

	if user >= 3 {
		user = 1
	} else {
		user++
	}
}

func executeFile(currentUser, file string, language Language) (string, string) {
	script := fmt.Sprintf("/sofi/languages/%s/run.sh", strings.ToLower(language.Name))

	run := exec.Command("/bin/bash", script, currentUser, file)
	head := exec.Command("head", "--bytes", maxOutputBufferCapacity)

	errBuffer := bytes.Buffer{}
	run.Stderr = &errBuffer

	head.Stdin, _ = run.StdoutPipe()
	headOutput := bytes.Buffer{}
	head.Stdout = &headOutput

	_ = run.Start()
	_ = head.Start()
	_ = run.Wait()
	_ = head.Wait()

	result := ""

	if headOutput.Len() > 0 {
		result = headOutput.String()
	} else if headOutput.Len() == 0 && errBuffer.Len() == 0 {
		result = headOutput.String()
	}

	return result, errBuffer.String()
}

func cleanProcesses(currentUser string) error {
	return exec.Command("pkill", "-9", "-u", currentUser).Run()
}

func restoreUserDir(currentUser string) {
	userDir := "/tmp/" + currentUser
	if _, err := ioutil.ReadDir(userDir); err != nil {
		if os.IsNotExist(err) {
			_ = exec.Command("runuser", "-u", currentUser, "--", "mkdir", userDir).Run()
		}
	}
}
