package awscli

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/test"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"go.uber.org/zap"
	"os/exec"
	"testing"
	"time"
)

func TestClient_Upload(t *testing.T) {
	cmdExecutor := &test.Executor{}
	client := NewClient(
		Options{
			Binary:      "aws",
			Bucket:      "test-bucket",
			EndpointUrl: "test-endpoint",
			Profile:     "test-profile",
		},
		zap.S(),
	)

	url, err := client.Upload(context.Background(), cmdExecutor, "source-dir", "dest-dir")
	require.NoError(t, err)
	require.Equal(t, "s3://test-bucket/dest-dir", url)
	require.Equal(
		t,
		"aws s3 sync source-dir 's3://test-bucket/dest-dir' --endpoint-url test-endpoint --profile test-profile",
		cmdExecutor.LastCmd.String(),
	)
}

func TestClient_ListBackups(t *testing.T) {
	testOutputs := []string{
		// an output of "aws s3 ls" at depth=0
		`
this is an output of "aws s3 ls"

                           PRE dir-1/
                           PRE dir 2 with spaces/
`,
		// an output of "aws s3 ls" at depth=1, for a "dir-1" directory from above.
		// an empty string means it doesn't contain any other directories.
		``,
		// an output of "aws s3 ls" at depth=1, for a "dir 2 with spaces" directory from above.
		// it contains one directory, common-scylla1-dc1.
		`PRE common-scylla1-dc1/`,
		// an output of "aws s3 ls" at depth=2, for a "dir 2 with spaces" directory.
		// it contains one directory, 09-07-2021-10-29, which should be recognized as a backup.
		`PRE 09-07-2021-10-29/`,
	}
	cmdExecutor := &test.Executor{
		Func: func(cmd *exec.Cmd, executedCount int) (string, error) {
			if len(testOutputs) > executedCount {
				return testOutputs[executedCount], nil
			} else {
				return "", nil
			}
		},
	}
	client := NewClient(
		Options{
			Bucket:      "test-bucket",
			EndpointUrl: "http://test-s3",
			Profile:     "test-profile",
		},
		zap.S(),
	)
	backups, err := client.ListBackups(context.Background(), cmdExecutor, "")
	require.NoError(t, err)
	require.Equal(
		t,
		[]entity.RemoteBackup{
			{
				Path:        "/dir 2 with spaces/common-scylla1-dc1/09-07-2021-10-29",
				HostPrefix:  "common-scylla1-dc1",
				DateCreated: time.Date(2021, 9, 7, 10, 29, 0, 0, time.UTC),
			},
		},
		backups,
	)
}
