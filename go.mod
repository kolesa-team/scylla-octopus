module github.com/kolesa-team/scylla-octopus

go 1.17

require (
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/hashicorp/go-multierror v1.1.1
	github.com/melbahja/goph v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/pkg/sftp v1.13.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20211020174200-9d6173849985 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/melbahja/goph v1.2.1 => github.com/antonsergeyev/goph v1.2.3
