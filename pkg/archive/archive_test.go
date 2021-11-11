package archive

import (
	"context"
	"fmt"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/local"
	"github.com/kolesa-team/scylla-octopus/pkg/entity"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func Test_PigzBinaryExists(t *testing.T) {
	executor := local.Executor{}
	ctx := context.Background()
	require.NoError(t, cmd.ExecutableFileExists(ctx, executor, "/bin/pigz"), "установите pigz")
}

func Test_Compress(t *testing.T) {
	cmdExecutor := local.Executor{}
	ctx := context.Background()
	node := entity.NewNode(entity.NewNodeInfo(
		"scylla.test",
		"/var/scylla/data",
		entity.NodeBinaries{
			Cqlsh: "cqlsh",
		},
		false,
	), cmdExecutor, nil)

	archive := entity.Archive{
		Method: "pigz",
		ArchiveOptions: entity.ArchiveOptions{
			Compression: "9",
			Threads:     "4",
		},
	}

	dir, err := ioutil.TempDir("", "backup")
	require.NoError(t, err)
	file1, err := ioutil.TempFile(dir, "file1.yml")
	require.NoError(t, err)
	file2, err := ioutil.TempFile(dir, "file2.yml")
	require.NoError(t, err)

	err = Compress(ctx, node, dir, archive)
	require.NoError(t, err)

	require.False(t, isFileExist(file1.Name()), fmt.Sprintf("file \"%s\" was not deleted", file1.Name()))
	require.False(t, isFileExist(file2.Name()), fmt.Sprintf("file \"%s\" was not deleted", file1.Name()))

	require.True(t, isFileExist(dir+"/backup.tar.pigz"), fmt.Sprintf("archive not found"))

	err = os.RemoveAll(dir)
	require.NoError(t, err)
}

func isFileExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}

	return true
}
