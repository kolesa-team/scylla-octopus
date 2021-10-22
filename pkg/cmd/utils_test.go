package cmd

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/kolesa-team/scylla-octopus/pkg/cmd/local"
	"os"
	"testing"
	"time"
)

func Test_BinaryExists(t *testing.T) {
	executor := local.Executor{}
	ctx := context.Background()
	require.NoError(t, ExecutableFileExists(ctx, executor, "/bin/ls"), "проверка наличия программы ls должна быть успешной")
	require.Error(t, ExecutableFileExists(ctx, executor, "ololo"), "проверка наличия программы ololo не должна быть успешной")
}

func Test_DirectoryExists(t *testing.T) {
	executor := local.Executor{}
	ctx := context.Background()
	require.True(t, DirectoryExists(ctx, executor, "/usr"))
	require.False(t, DirectoryExists(ctx, executor, "/ololo"))
}

func Test_CreateDirectory(t *testing.T) {
	executor := local.Executor{}
	ctx := context.Background()
	path := fmt.Sprintf("%s/%d", os.TempDir(), time.Now().UnixNano())
	require.False(t, DirectoryExists(ctx, executor, path))

	require.NoError(t, CreateDirectory(ctx, executor, path), "директория должна быть создана без ошибок")

	require.True(t, DirectoryExists(ctx, executor, path), "директория должна существовать")

	require.NoError(t, RemoveDirectory(ctx, executor, path), "директория должна быть удалена без ошибок")

	require.False(t, DirectoryExists(ctx, executor, path), "директория не должна существовать")
}

// Проверка отмены выполнения команды.
// Ожидается, что тест завершится быстро, а команда sleep 10 будет прервана.
func Test_CtxCancel(t *testing.T) {
	executor := local.Executor{}
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	resultChan := make(chan struct{})

	go func() {
		err = executor.Run(ctx, Command("sleep", "10"))
		resultChan <- struct{}{}
	}()

	cancel()
	<-resultChan

	require.Error(t, err, "контекст должен быть отменён")
}
