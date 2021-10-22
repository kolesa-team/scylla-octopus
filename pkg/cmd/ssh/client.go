package ssh

// This package manages SSH connections to remote hosts and implements cmd.Executor.
// Based on a fork of github.com/melbahja/goph with context cancellation support (github.com/antonsergeyev/goph).

import (
	"bytes"
	"context"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"os/exec"
	"sync"
	"time"
)

type Options struct {
	Port        uint
	Username    string
	Password    string
	KeyFile     string `yaml:"keyFile"`
	KeyPassword string `yaml:"keyPassword"`
	Debug       bool
}

type Client struct {
	options   Options
	executors sync.Map
	logger    *zap.SugaredLogger
	auth      goph.Auth
}

// HostExecutor implements a shell command executor on a remote machine over SSH
type HostExecutor struct {
	host    string
	debug   bool
	sshConn *goph.Client
	logger  *zap.SugaredLogger
}

func NewClient(opts Options, logger *zap.SugaredLogger) (client *Client, err error) {
	if opts.Port == 0 {
		opts.Port = 22
	}

	client = &Client{
		options:   opts,
		logger:    logger,
		executors: sync.Map{},
	}

	if len(opts.KeyFile) > 0 {
		client.auth, err = goph.Key(opts.KeyFile, opts.KeyPassword)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"could not read SSH key from %s",
				opts.KeyFile,
			)
		}
	} else {
		client.auth = goph.Password(opts.Password)
	}

	return client, nil
}

// GetByHost returns a shell command executor on a remote host.
// Attempts to only create SSH connections once, keeping them in a `sync.Map`.
func (c *Client) GetByHost(host string) (*HostExecutor, error) {
	_, ok := c.executors.Load(host)
	if !ok {
		logger := c.logger.With("host", host)

		logger.Debugw(
			"creating SSH connection",
			"host",
			host,
		)
		sshConn, err := goph.NewConn(&goph.Config{
			Auth:     c.auth,
			User:     c.options.Username,
			Addr:     host,
			Port:     c.options.Port,
			Timeout:  time.Second * 2,
			Callback: ssh.InsecureIgnoreHostKey(),
		})

		if err != nil {
			return nil, errors.Wrapf(
				err,
				"could not create SSH connection to %s as %s",
				host,
				c.options.Username,
			)
		}

		c.executors.Store(host, &HostExecutor{
			host:    host,
			debug:   c.options.Debug,
			sshConn: sshConn,
			logger:  logger,
		})
	}

	result, _ := c.executors.Load(host)

	return result.(*HostExecutor), nil
}

func (h *HostExecutor) Execute(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	timeStarted := time.Now()

	if h.debug {
		fmt.Printf("\n---[SSH] executing command at %s:---\n%s\n", h.host, cmd.String())
	}

	sshCmd, err := h.sshConn.CommandContext(ctx, cmd.String())
	if err != nil {
		return nil, err
	}

	output, err := sshCmd.CombinedOutput()

	if h.debug {
		fmt.Printf(
			"\n---[SSH] command at %s done in %s, output:---\n%s\n",
			h.host,
			time.Now().Sub(timeStarted).String(),
			string(output),
		)
	}

	return output, err
}

func (h *HostExecutor) Run(ctx context.Context, cmd *exec.Cmd) error {
	_, err := h.Execute(ctx, cmd)
	return err
}

func (h *HostExecutor) ReadFile(ctx context.Context, path string) ([]byte, error) {
	var data bytes.Buffer
	err := h.sshConn.ReadFile(path, &data)

	return data.Bytes(), err
}

func (h *HostExecutor) WriteFile(ctx context.Context, path string, data []byte) error {
	source := bytes.NewReader(data)

	return h.sshConn.WriteFile(path, source)
}
