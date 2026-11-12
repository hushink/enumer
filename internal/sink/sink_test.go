package sink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "file.txt")

	changed, err := Write(path, []byte("one"))
	require.NoError(t, err)
	require.True(t, changed)

	changed, err = Write(path, []byte("one"))
	require.NoError(t, err)
	require.False(t, changed)

	changed, err = Write(path, []byte("two"))
	require.NoError(t, err)
	require.True(t, changed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "two", string(data))
}
