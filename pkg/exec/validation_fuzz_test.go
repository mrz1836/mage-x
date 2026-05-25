package exec

import (
	"errors"
	"strings"
	"testing"
)

// FuzzValidateCommandArg fuzzes the command-arg validator with arbitrary input.
// The validator must never panic, never accept clearly dangerous shell metacharacters,
// and must reject inputs containing $(, backtick, &&, ||, ;, >, < unconditionally.
func FuzzValidateCommandArg(f *testing.F) {
	seeds := []string{
		"",
		"hello",
		"/path/to/file",
		"--verbose",
		"https://example.com/path?q=1|2",
		"^test|foo$",
		"$(whoami)",
		"`whoami`",
		"foo && bar",
		"foo || bar",
		"foo; bar",
		"foo > bar",
		"foo < bar",
		"test | cat",
		"\x00null",
		"\xff\xfe invalid utf8",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	mustReject := []string{"$(", "`", "&&", "||", ";", ">", "<"}

	f.Fuzz(func(t *testing.T, arg string) {
		err := ValidateCommandArg(arg)
		for _, bad := range mustReject {
			if strings.Contains(arg, bad) && err == nil {
				t.Fatalf("argument %q contains %q but was accepted", arg, bad)
			}
		}
		if err != nil && !errors.Is(err, ErrDangerousPattern) &&
			!errors.Is(err, ErrDangerousPipePattern) &&
			!errors.Is(err, ErrInvalidUTF8) {
			t.Fatalf("unexpected error type for %q: %v", arg, err)
		}
	})
}
