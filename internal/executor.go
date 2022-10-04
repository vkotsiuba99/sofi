package internal

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/exec"
	"sofi/internal/pool"
	"strings"
)

const (
	maxOutputBufferCapacity = "65332"
)

type RceEngine struct {
	systemUsers *pool.SystemUsers
	pool        *pool.WorkerPool
}

func NewRceEngine() *RceEngine {
	return &RceEngine{
		systemUsers: pool.NewSystemUser(50),
		pool:        pool.NewWorkerPool(50, 100),
	}
}

func (rce *RceEngine) action(lang, code string, ch chan<- pool.CodeOutput) {
	language, err := GetLanguageByName(lang)
	if err != nil {
		ch <- pool.CodeOutput{}
		return
	}

	user, err := rce.systemUsers.Acquire()
	if err != nil {
		rce.systemUsers.Release(user.Uid)
		ch <- pool.CodeOutput{}
		return
	}

	tempDirName := uuid.New().String()

	err = CreateTempDir(user.Username, tempDirName)
	if err != nil {
		rce.systemUsers.Release(user.Uid)
		ch <- pool.CodeOutput{}
		return
	}

	filename, err := CreateTempFile(user.Username, tempDirName, language.Extension)
	if err != nil {
		rce.systemUsers.Release(user.Uid)
		DeleteTempDir(user.Username, tempDirName)
		ch <- pool.CodeOutput{}
		return
	}

	err = WriteToFile(filename, code)
	if err != nil {
		rce.systemUsers.Release(user.Uid)
		ch <- pool.CodeOutput{}
		return
	}

	output, errorString := rce.executeFile(user.Username, filename, language)

	ch <- pool.CodeOutput{
		User:        *user,
		TempDirName: tempDirName,
		Result:      output,
		Error:       errorString,
	}

	rce.CleanUp(user, tempDirName)
}

func (rce *RceEngine) Dispatch(lang, code string) (pool.CodeOutput, error) {
	dataChannel := make(chan pool.CodeOutput)
	rce.pool.SubmitJob(lang, code, rce.action, dataChannel)
	output := <-dataChannel
	return output, nil
}

func (rce *RceEngine) CleanUp(user *pool.User, tempDirName string) {
	DeleteTempDir(user.Username, tempDirName)
	rce.cleanProcesses(user.Username)
	rce.restoreUserDir(user.Username)
	rce.systemUsers.Release(user.Uid)
}

func (rce *RceEngine) executeFile(currentUser, file string, language Language) (string, string) {
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

func (rce *RceEngine) cleanProcesses(currentUser string) error {
	return exec.Command("pkill", "-9", "-u", currentUser).Run()
}

func (rce *RceEngine) restoreUserDir(currentUser string) {
	userDir := "/tmp/" + currentUser
	if _, err := os.ReadDir(userDir); err != nil {
		if os.IsNotExist(err) {
			_ = exec.Command("runuser", "-u", currentUser, "--", "mkdir", userDir).Run()
		}
	}
}
