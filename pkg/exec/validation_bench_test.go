package exec

import "testing"

// BenchmarkValidateCommandArg measures the cost of arg validation for typical
// well-formed inputs. Validation runs on every command arg, so allocation-free
// fast paths matter.
func BenchmarkValidateCommandArg(b *testing.B) {
	args := []string{
		"hello",
		"/path/to/file",
		"--verbose",
		"https://example.com/path?q=1|2",
		"^test|foo$",
		"argument-with-many-hyphens-and-numbers-12345",
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		if err := ValidateCommandArg(args[i%len(args)]); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

// BenchmarkValidateCommandArg_Rejected measures the rejection path for inputs
// that contain dangerous shell metacharacters.
func BenchmarkValidateCommandArg_Rejected(b *testing.B) {
	args := []string{
		"$(whoami)",
		"`whoami`",
		"foo && bar",
		"foo || bar",
		"foo; bar",
		"foo > out",
		"test | cat",
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		if err := ValidateCommandArg(args[i%len(args)]); err == nil {
			b.Fatal("expected validation error")
		}
		i++
	}
}
