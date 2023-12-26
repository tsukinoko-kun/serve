package utils_test

import (
	"testing"

	"github.com/Frank-Mayer/serve/internal/utils"
)

func isInTestWrapper(t *testing.T, path string, dir string, expected bool) {
	result := utils.IsIn(path, dir)
	if result != expected {
		t.Errorf("IsIn(%s, %s) = %t, expected %t", path, dir, result, expected)
	}
}

func TestIsIn(t *testing.T) {
	t.Parallel()
	isInTestWrapper(t, "/home/user/.config", "/home/user", true)
	isInTestWrapper(t, "/home/user/.config", "/home/user/", true)
	isInTestWrapper(t, "/home/user/.config", "/home/user/.config", true)
	isInTestWrapper(t, "/home/user/.config", "/home/user/.config/", true)
	isInTestWrapper(t, "/home/user/.config", "/home/user/.config/.", true)
	isInTestWrapper(t, "/home/user/.config", "/home/user/.config/..", true)
	isInTestWrapper(t, "/home/user/.config", "/home/user/.config/../../", true)
	isInTestWrapper(t, "/home/user/.config/..", "/home/user/", true)
	isInTestWrapper(t, "./a/b/c", "./a/b", true)
	isInTestWrapper(t, "./a/b/c/../d/./e", "./a/b/c/..", true)

    // false cases
	isInTestWrapper(t, "./a", "./a/b/c/..", false)
    isInTestWrapper(t, "./a/b/c/..", "./d", false)
    isInTestWrapper(t, "./a/b/c/../..", "./a/b/c/..", false)
    isInTestWrapper(t, "/a/b/c/../..", "/a/b/c/..", false)
}
