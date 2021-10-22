package factory

import (
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/local"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/ssh"
	"go.uber.org/zap"
)

// Factory of shell commands executors.
// Depending on the options, will create local or SSH executors.
type Factory struct {
	options   Options
	sshClient *ssh.Client
}

type Options struct {
	UseSSH bool
	SSH    ssh.Options `yaml:"ssh"`
	Debug  bool
}

// GetByHost returns a command executor for a given host.
func (f Factory) GetByHost(host string) (cmd.Executor, error) {
	if !f.options.UseSSH {
		return local.Executor{
			Debug: f.options.Debug,
		}, nil
	}

	return f.sshClient.GetByHost(host)
}

func NewFactory(opts Options, logger *zap.SugaredLogger) (Factory, error) {
	factory := Factory{options: opts}
	var err error

	if opts.UseSSH {
		opts.SSH.Debug = opts.Debug
		factory.sshClient, err = ssh.NewClient(opts.SSH, logger)
	}

	return factory, err
}

// NewTestFactory creates a factory for testing,
// that only returns local command executors.
func NewTestFactory() Factory {
	return Factory{
		options: Options{
			Debug: true,
		},
	}
}
