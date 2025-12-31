package fileops

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Static errors for mock tests.
var (
	errMockChmod     = errors.New("mock chmod error")
	errMockReadDir   = errors.New("mock readdir error")
	errMockReadFile  = errors.New("mock readfile error")
	errMockRemove    = errors.New("mock remove error")
	errMockRemoveAll = errors.New("mock removeall error")
	errMockStat      = errors.New("mock stat error")
	errMockWriteFile = errors.New("mock writefile error")
	errMockMarshal   = errors.New("mock marshal error")
	errMockUnmarshal = errors.New("mock unmarshal error")
	errMockWriteJSON = errors.New("mock writejson error")
	errMockWriteYAML = errors.New("mock writeyaml error")
)

// mockFileInfo implements fs.FileInfo for testing.
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// mockDirEntry implements fs.DirEntry for testing.
type mockDirEntry struct {
	name  string
	isDir bool
	mode  os.FileMode
	info  fs.FileInfo
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return m.mode }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return m.info, nil }

// TestMockFileOperatorChmod tests the MockFileOperator Chmod method.
func TestMockFileOperatorChmod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().Chmod("/test/file", os.FileMode(0o644)).Return(nil)

		err := mock.Chmod("/test/file", 0o644)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().Chmod("/test/file", os.FileMode(0o755)).Return(errMockChmod)

		err := mock.Chmod("/test/file", 0o755)
		require.Error(t, err)
		assert.Equal(t, errMockChmod, err)
	})
}

// TestMockFileOperatorIsDir tests the MockFileOperator IsDir method.
func TestMockFileOperatorIsDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("is_directory", func(t *testing.T) {
		mock.EXPECT().IsDir("/test/dir").Return(true)

		result := mock.IsDir("/test/dir")
		assert.True(t, result)
	})

	t.Run("not_directory", func(t *testing.T) {
		mock.EXPECT().IsDir("/test/file.txt").Return(false)

		result := mock.IsDir("/test/file.txt")
		assert.False(t, result)
	})
}

// TestMockFileOperatorReadDir tests the MockFileOperator ReadDir method.
func TestMockFileOperatorReadDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		entries := []fs.DirEntry{
			&mockDirEntry{name: "file1.txt", isDir: false},
			&mockDirEntry{name: "subdir", isDir: true},
		}
		mock.EXPECT().ReadDir("/test/dir").Return(entries, nil)

		result, err := mock.ReadDir("/test/dir")
		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "file1.txt", result[0].Name())
		assert.Equal(t, "subdir", result[1].Name())
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().ReadDir("/nonexistent").Return(nil, errMockReadDir)

		result, err := mock.ReadDir("/nonexistent")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, errMockReadDir, err)
	})
}

// TestMockFileOperatorReadFile tests the MockFileOperator ReadFile method.
func TestMockFileOperatorReadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		expectedData := []byte("file contents")
		mock.EXPECT().ReadFile("/test/file.txt").Return(expectedData, nil)

		data, err := mock.ReadFile("/test/file.txt")
		require.NoError(t, err)
		assert.Equal(t, expectedData, data)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().ReadFile("/nonexistent").Return(nil, errMockReadFile)

		data, err := mock.ReadFile("/nonexistent")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, errMockReadFile, err)
	})
}

// TestMockFileOperatorRemove tests the MockFileOperator Remove method.
func TestMockFileOperatorRemove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().Remove("/test/file.txt").Return(nil)

		err := mock.Remove("/test/file.txt")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().Remove("/protected/file").Return(errMockRemove)

		err := mock.Remove("/protected/file")
		require.Error(t, err)
		assert.Equal(t, errMockRemove, err)
	})
}

// TestMockFileOperatorRemoveAll tests the MockFileOperator RemoveAll method.
func TestMockFileOperatorRemoveAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().RemoveAll("/test/dir").Return(nil)

		err := mock.RemoveAll("/test/dir")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().RemoveAll("/protected/dir").Return(errMockRemoveAll)

		err := mock.RemoveAll("/protected/dir")
		require.Error(t, err)
		assert.Equal(t, errMockRemoveAll, err)
	})
}

// TestMockFileOperatorStat tests the MockFileOperator Stat method.
func TestMockFileOperatorStat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		info := &mockFileInfo{
			name:    "test.txt",
			size:    1024,
			mode:    0o644,
			modTime: time.Now(),
			isDir:   false,
		}
		mock.EXPECT().Stat("/test/file.txt").Return(info, nil)

		result, err := mock.Stat("/test/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "test.txt", result.Name())
		assert.Equal(t, int64(1024), result.Size())
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().Stat("/nonexistent").Return(nil, errMockStat)

		result, err := mock.Stat("/nonexistent")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, errMockStat, err)
	})
}

// TestMockFileOperatorWriteFile tests the MockFileOperator WriteFile method.
func TestMockFileOperatorWriteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := []byte("test data")
		mock.EXPECT().WriteFile("/test/file.txt", data, os.FileMode(0o644)).Return(nil)

		err := mock.WriteFile("/test/file.txt", data, 0o644)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := []byte("test data")
		mock.EXPECT().WriteFile("/readonly/file.txt", data, os.FileMode(0o644)).Return(errMockWriteFile)

		err := mock.WriteFile("/readonly/file.txt", data, 0o644)
		require.Error(t, err)
		assert.Equal(t, errMockWriteFile, err)
	})
}

// TestMockJSONOperatorMarshal tests the MockJSONOperator Marshal method.
func TestMockJSONOperatorMarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockJSONOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		expectedJSON := []byte(`{"key":"value"}`)
		mock.EXPECT().Marshal(data).Return(expectedJSON, nil)

		result, err := mock.Marshal(data)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), string(result))
	})

	t.Run("error", func(t *testing.T) {
		data := make(chan int) // Unmarshalable type
		mock.EXPECT().Marshal(data).Return(nil, errMockMarshal)

		result, err := mock.Marshal(data)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, errMockMarshal, err)
	})
}

// TestMockJSONOperatorUnmarshal tests the MockJSONOperator Unmarshal method.
func TestMockJSONOperatorUnmarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockJSONOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := []byte(`{"key":"value"}`)
		var target map[string]string
		mock.EXPECT().Unmarshal(data, &target).Return(nil)

		err := mock.Unmarshal(data, &target)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := []byte(`invalid json`)
		var target map[string]string
		mock.EXPECT().Unmarshal(data, &target).Return(errMockUnmarshal)

		err := mock.Unmarshal(data, &target)
		require.Error(t, err)
		assert.Equal(t, errMockUnmarshal, err)
	})
}

// TestMockJSONOperatorWriteJSON tests the MockJSONOperator WriteJSON method.
func TestMockJSONOperatorWriteJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockJSONOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		mock.EXPECT().WriteJSON("/test/file.json", data).Return(nil)

		err := mock.WriteJSON("/test/file.json", data)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		mock.EXPECT().WriteJSON("/readonly/file.json", data).Return(errMockWriteJSON)

		err := mock.WriteJSON("/readonly/file.json", data)
		require.Error(t, err)
		assert.Equal(t, errMockWriteJSON, err)
	})
}

// TestMockJSONOperatorWriteJSONIndent tests the MockJSONOperator WriteJSONIndent method.
func TestMockJSONOperatorWriteJSONIndent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockJSONOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		mock.EXPECT().WriteJSONIndent("/test/file.json", data, "", "  ").Return(nil)

		err := mock.WriteJSONIndent("/test/file.json", data, "", "  ")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		mock.EXPECT().WriteJSONIndent("/readonly/file.json", data, "", "  ").Return(errMockWriteJSON)

		err := mock.WriteJSONIndent("/readonly/file.json", data, "", "  ")
		require.Error(t, err)
		assert.Equal(t, errMockWriteJSON, err)
	})
}

// TestMockYAMLOperatorUnmarshal tests the MockYAMLOperator Unmarshal method.
func TestMockYAMLOperatorUnmarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockYAMLOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := []byte("key: value")
		var target map[string]string
		mock.EXPECT().Unmarshal(data, &target).Return(nil)

		err := mock.Unmarshal(data, &target)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := []byte("invalid: yaml: content")
		var target map[string]string
		mock.EXPECT().Unmarshal(data, &target).Return(errMockUnmarshal)

		err := mock.Unmarshal(data, &target)
		require.Error(t, err)
		assert.Equal(t, errMockUnmarshal, err)
	})
}

// TestMockYAMLOperatorWriteYAML tests the MockYAMLOperator WriteYAML method.
func TestMockYAMLOperatorWriteYAML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockYAMLOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		mock.EXPECT().WriteYAML("/test/file.yaml", data).Return(nil)

		err := mock.WriteYAML("/test/file.yaml", data)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		mock.EXPECT().WriteYAML("/readonly/file.yaml", data).Return(errMockWriteYAML)

		err := mock.WriteYAML("/readonly/file.yaml", data)
		require.Error(t, err)
		assert.Equal(t, errMockWriteYAML, err)
	})
}

// TestMockSafeFileOperatorWriteFileAtomic tests MockSafeFileOperator WriteFileAtomic.
func TestMockSafeFileOperatorWriteFileAtomic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := []byte("atomic data")
		mock.EXPECT().WriteFileAtomic("/test/file.txt", data, os.FileMode(0o644)).Return(nil)

		err := mock.WriteFileAtomic("/test/file.txt", data, 0o644)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := []byte("atomic data")
		mock.EXPECT().WriteFileAtomic("/readonly/file.txt", data, os.FileMode(0o644)).Return(errMockWriteFile)

		err := mock.WriteFileAtomic("/readonly/file.txt", data, 0o644)
		require.Error(t, err)
		assert.Equal(t, errMockWriteFile, err)
	})
}

// TestMockSafeFileOperatorWriteFileWithBackup tests MockSafeFileOperator WriteFileWithBackup.
func TestMockSafeFileOperatorWriteFileWithBackup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := []byte("backup data")
		mock.EXPECT().WriteFileWithBackup("/test/file.txt", data, os.FileMode(0o644)).Return(nil)

		err := mock.WriteFileWithBackup("/test/file.txt", data, 0o644)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := []byte("backup data")
		mock.EXPECT().WriteFileWithBackup("/readonly/file.txt", data, os.FileMode(0o644)).Return(errMockWriteFile)

		err := mock.WriteFileWithBackup("/readonly/file.txt", data, 0o644)
		require.Error(t, err)
		assert.Equal(t, errMockWriteFile, err)
	})
}

// TestMockSafeFileOperatorChmod tests MockSafeFileOperator Chmod.
func TestMockSafeFileOperatorChmod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().Chmod("/test/file", os.FileMode(0o755)).Return(nil)

		err := mock.Chmod("/test/file", 0o755)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().Chmod("/protected/file", os.FileMode(0o755)).Return(errMockChmod)

		err := mock.Chmod("/protected/file", 0o755)
		require.Error(t, err)
		assert.Equal(t, errMockChmod, err)
	})
}

// TestMockSafeFileOperatorIsDir tests MockSafeFileOperator IsDir.
func TestMockSafeFileOperatorIsDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("is_directory", func(t *testing.T) {
		mock.EXPECT().IsDir("/test/dir").Return(true)

		result := mock.IsDir("/test/dir")
		assert.True(t, result)
	})

	t.Run("not_directory", func(t *testing.T) {
		mock.EXPECT().IsDir("/test/file.txt").Return(false)

		result := mock.IsDir("/test/file.txt")
		assert.False(t, result)
	})
}

// TestMockSafeFileOperatorReadDir tests MockSafeFileOperator ReadDir.
func TestMockSafeFileOperatorReadDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		entries := []fs.DirEntry{
			&mockDirEntry{name: "file.txt", isDir: false},
		}
		mock.EXPECT().ReadDir("/test/dir").Return(entries, nil)

		result, err := mock.ReadDir("/test/dir")
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "file.txt", result[0].Name())
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().ReadDir("/nonexistent").Return(nil, errMockReadDir)

		result, err := mock.ReadDir("/nonexistent")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestMockSafeFileOperatorReadFile tests MockSafeFileOperator ReadFile.
func TestMockSafeFileOperatorReadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		expected := []byte("file data")
		mock.EXPECT().ReadFile("/test/file.txt").Return(expected, nil)

		result, err := mock.ReadFile("/test/file.txt")
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().ReadFile("/nonexistent").Return(nil, errMockReadFile)

		result, err := mock.ReadFile("/nonexistent")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestMockSafeFileOperatorRemoveAll tests MockSafeFileOperator RemoveAll.
func TestMockSafeFileOperatorRemoveAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().RemoveAll("/test/dir").Return(nil)

		err := mock.RemoveAll("/test/dir")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().RemoveAll("/protected").Return(errMockRemoveAll)

		err := mock.RemoveAll("/protected")
		require.Error(t, err)
		assert.Equal(t, errMockRemoveAll, err)
	})
}

// TestMockSafeFileOperatorStat tests MockSafeFileOperator Stat.
func TestMockSafeFileOperatorStat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		info := &mockFileInfo{name: "test.txt", size: 100}
		mock.EXPECT().Stat("/test/file.txt").Return(info, nil)

		result, err := mock.Stat("/test/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "test.txt", result.Name())
		assert.Equal(t, int64(100), result.Size())
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().Stat("/nonexistent").Return(nil, errMockStat)

		result, err := mock.Stat("/nonexistent")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestMockSafeFileOperatorWriteFile tests MockSafeFileOperator WriteFile.
func TestMockSafeFileOperatorWriteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockSafeFileOperator(ctrl)

	t.Run("success", func(t *testing.T) {
		data := []byte("test data")
		mock.EXPECT().WriteFile("/test/file.txt", data, os.FileMode(0o644)).Return(nil)

		err := mock.WriteFile("/test/file.txt", data, 0o644)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		data := []byte("test data")
		mock.EXPECT().WriteFile("/readonly/file.txt", data, os.FileMode(0o644)).Return(errMockWriteFile)

		err := mock.WriteFile("/readonly/file.txt", data, 0o644)
		require.Error(t, err)
		assert.Equal(t, errMockWriteFile, err)
	})
}

// TestFileOpsWithMockedDependencies tests FileOps with mocked dependencies.
func TestFileOpsWithMockedDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFile := NewMockFileOperator(ctrl)
	mockJSON := NewMockJSONOperator(ctrl)
	mockYAML := NewMockYAMLOperator(ctrl)
	mockSafe := NewMockSafeFileOperator(ctrl)

	ops := &FileOps{
		File: mockFile,
		JSON: mockJSON,
		YAML: mockYAML,
		Safe: mockSafe,
	}

	t.Run("WriteYAMLSafe_marshal_error", func(t *testing.T) {
		mockFile.EXPECT().Exists(gomock.Any()).Return(true)
		mockYAML.EXPECT().Marshal(gomock.Any()).Return(nil, errMockMarshal)

		err := ops.WriteYAMLSafe("/test/file.yaml", map[string]string{"key": "value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal YAML")
	})

	t.Run("CleanupBackups_remove_error", func(t *testing.T) {
		// This test uses the actual glob, so we need a real directory
		tmpDir := t.TempDir()

		// Create a backup file to match the pattern
		backupPath := tmpDir + "/test.bak"
		require.NoError(t, os.WriteFile(backupPath, []byte("backup"), 0o644)) //nolint:gosec // G306: Test file

		mockFile.EXPECT().Remove(backupPath).Return(errMockRemove)

		err := ops.CleanupBackups(tmpDir, "*.bak")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove backup")
	})
}
