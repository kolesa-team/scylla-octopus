package environment

import (
	"github.com/kolesa-team/scylla-octopus/app"
	"github.com/kolesa-team/scylla-octopus/app/backup"
	"github.com/kolesa-team/scylla-octopus/pkg/awscli"
	"github.com/kolesa-team/scylla-octopus/pkg/cluster"
	cmdFactory "github.com/kolesa-team/scylla-octopus/pkg/cmd/factory"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"github.com/kolesa-team/scylla-octopus/pkg/scylla"
	"go.uber.org/zap"
)

// Environment holds all the project dependencies
type Environment struct {
	Config        Config
	BuildInfo     entity.BuildInfo
	Logger        *zap.SugaredLogger
	Scylla        *scylla.Client
	AwsCli        *awscli.Client
	CmdFactory    cmdFactory.Factory
	Cluster       *cluster.Cluster
	Notifier      notifier.Notifier
	BackupService *backup.Service
	App           *app.Octopus
}

// GetEnvironment initializes project dependencies
func GetEnvironment(cfg Config, buildInfo entity.BuildInfo) (Environment, error) {
	var err error
	env := Environment{
		Config:    cfg,
		BuildInfo: buildInfo,
		Logger:    getLogger(cfg.Log),
	}

	env.Notifier = notifier.New(cfg.Notifier, env.Logger)
	env.Scylla = scylla.NewClient(cfg.Credentials, env.Logger)

	if cfg.Awscli == nil {
		// if awscli is not configured, create empty options so that we don't have to deal with nil or interfaces
		cfg.Awscli = &awscli.Options{Disabled: true}
	}

	env.AwsCli = awscli.NewClient(*cfg.Awscli, env.Logger)

	env.CmdFactory, err = cmdFactory.NewFactory(cfg.Commands, env.Logger)
	if err != nil {
		return env, err
	}

	env.Cluster = cluster.NewCluster(
		cfg.Cluster,
		env.CmdFactory,
		env.Logger,
	)

	env.BackupService = backup.NewService(
		cfg.Backup,
		buildInfo,
		env.Scylla,
		env.AwsCli,
		env.Notifier,
		env.Logger,
	)

	env.App = app.NewOctopus(
		env.Cluster,
		env.Scylla,
		env.BackupService,
		env.AwsCli,
		env.Notifier,
		env.Logger,
	)

	return env, nil
}
