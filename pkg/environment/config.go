package environment

import (
	"github.com/go-yaml/yaml"
	"github.com/kolesa-team/scylla-octopus/app/backup"
	"github.com/kolesa-team/scylla-octopus/pkg/awscli"
	"github.com/kolesa-team/scylla-octopus/pkg/cluster"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/factory"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/kolesa-team/scylla-octopus/pkg/notifier"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net"
)

// Config holds project configuration
type Config struct {
	Cluster     cluster.Options
	Credentials entity.Credentials
	Awscli      *awscli.Options
	Log         LogOptions
	Backup      backup.Options
	Notifier    notifier.Options
	Commands    factory.Options
}

func GetConfig(file string, forceVerboseMode bool) (Config, error) {
	cfg := Config{}
	cfgBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return cfg, errors.Wrap(err, "could not read configuration file")
	}

	err = yaml.Unmarshal(cfgBytes, &cfg)

	if err != nil {
		return cfg, errors.Wrap(err, "could not parse configuration file")
	}

	// sanity checks
	if cfg.Awscli == nil {
		if cfg.Backup.DisableUpload == false {
			return cfg, errors.New("awscli configuration is required if backup.disableUpload=false")
		}
	}

	if cfg.Backup.DisableUpload {
		if cfg.Backup.CleanupLocal == true {
			return cfg, errors.New("backup.cleanupLocal cannot be true if remote upload is disabled")
		}

		if cfg.Backup.CleanupRemote == true {
			return cfg, errors.New("backup.cleanupRemote cannot be true if remote upload is disabled")
		}
	}

	if forceVerboseMode {
		cfg.Log.Level = zapcore.DebugLevel
		cfg.Commands.Debug = true
	}

	if len(cfg.Cluster.Hosts) == 0 {
		// if no remote hosts are present, assume we're running locally
		cfg.Commands.UseSSH = false
		cfg.Cluster.Hosts = []string{getLocalIP()}
	} else {
		cfg.Commands.UseSSH = true
	}

	return cfg, nil
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}
