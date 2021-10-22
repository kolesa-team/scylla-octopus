package ssh

import (
	"context"
	"fmt"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"go.uber.org/zap"
	"strings"
	"time"
)

// An example usage of SSH client.
// Remove an "_" from function name to run.

// A database host from remote.yml
const exampleSshHost = "10.5.0.2"

func _ExampleSSHClient() {
	logger, _ := zap.NewDevelopment()
	opts := Options{
		Username: "root",
		KeyFile:  "./../../../test/ssh/id_rsa",
		Debug:    false,
	}
	client, err := NewClient(opts, logger.Sugar())
	if err != nil {
		logger.Fatal(err.Error())
	}

	executor, err := client.GetByHost(exampleSshHost)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// execute a `pwd` command and expect that we're in "/root" directory
	pwdOutput, err := executor.Execute(context.Background(), cmd.Command("pwd"))
	if err != nil {
		logger.Fatal("команда pwd должна быть выполнена успешно. получили ошибку " + err.Error())
	}
	pwdOutputLines := strings.Split(strings.TrimSpace(string(pwdOutput)), "\n")
	if pwdOutputLines[len(pwdOutputLines)-1] != "/root" {
		logger.Fatal("`pwd` expected to return /root, got " + pwdOutputLines[len(pwdOutputLines)-1])
	}

	// execute a `sleep 5` with a timeout and expect it to be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*20)
	defer cancel()
	_, err = executor.Execute(ctx, cmd.Command("sleep", "5"))

	if err == nil {
		logger.Fatal("error must not be nil")
	}

	fmt.Println("error: " + err.Error())

	// Output: error: context deadline exceeded
}
