package entity

import (
	"gopkg.in/yaml.v3"
	"time"
)

// BackupMetadata is written along with each backup, that will help us implement the restoration in the future.
type BackupMetadata struct {
	DateCreated time.Time `yaml:"dateCreated"`
	Host        string
	Keyspaces   []string
	SnapshotTag string    `yaml:"snapshotTag"`
	BuildInfo   BuildInfo `yaml:"buildInfo"`
	Archive     Archive   `yaml:"archive"`
}

func (b BackupMetadata) Bytes() []byte {
	data, _ := yaml.Marshal(b)

	return data
}
