package mage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode"

	pkglog "github.com/mrz1836/mage-x/pkg/log"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// errBumpJSONWithModule is returned when json is combined with module: the
// JSON outcome describes the root module's version tag.
var errBumpJSONWithModule = errors.New("json can't be combined with module: it reports the root module's version tag")

// bumpResult is the outcome of a version:bump run. With json, magex prints it
// as one JSON object on the last line of stdout, on success and on failure.
// Scripts read these field names, so they must not change. A step the run did
// not reach leaves its fields empty.
type bumpResult struct {
	From        string   `json:"from"`        // the version bumped from, as vX.Y.Z
	To          string   `json:"to"`          // the version bumped to, or in dry-run the one that would be
	Bump        string   `json:"bump"`        // patch, minor, or major
	Expect      string   `json:"expect"`      // the version expect pinned, as vX.Y.Z, or ""
	DryRun      bool     `json:"dryRun"`      // whether this was a dry-run
	Tag         string   `json:"tag"`         // the tag created, or ""
	Commit      string   `json:"commit"`      // the full SHA of the commit the tag points at, or ""
	Pushed      bool     `json:"pushed"`      // true only if origin has the tag at that commit
	StartBranch string   `json:"startBranch"` // the branch the clone was on, or "" when detached
	Branch      string   `json:"branch"`      // the branch the clone ended on, or "" when detached
	Warnings    []string `json:"warnings"`    // the warnings printed, as plain text; never null
	Error       string   `json:"error"`       // the error the run failed with, or ""
}

// bumpReport collects a version:bump run's outcome for json. A nil *bumpReport
// stands for a run without json and ignores every call, so the bump reads the
// same either way.
type bumpReport struct {
	result   bumpResult
	logger   pkglog.CLILogger // the CLI logger the recorder stands in for
	recorder *warningRecorder // keeps the warnings while the run prints them
}

// startBumpReport returns a bumpReport when args, or with none the arguments
// magex passes a delegated target, ask for json, and nil otherwise. It notes
// the branch the clone is on before the bump switches anywhere, and starts
// keeping the warnings the run prints.
func startBumpReport(args []string) *bumpReport {
	if len(args) == 0 {
		args = utils.MageArgs()
	}
	params := utils.ParseParams(args)
	if !utils.IsParamTrue(params, "json") {
		return nil
	}

	r := &bumpReport{
		result: bumpResult{
			DryRun:      utils.IsParamTrue(params, "dry-run"),
			StartBranch: currentBranchName(),
			Warnings:    []string{},
		},
	}
	if logger := pkglog.Default(); logger != nil {
		r.logger = logger
		r.recorder = &warningRecorder{CLILogger: logger}
		pkglog.SetDefault(r.recorder)
	}
	return r
}

// configure records the bump type and the expected version, and refuses
// module, whose tags the outcome does not describe.
func (r *bumpReport) configure(cfg *bumpConfig) error {
	if r == nil {
		return nil
	}
	r.result.Bump = cfg.bumpType
	r.result.Expect = cfg.expect
	if cfg.module != "" {
		return errBumpJSONWithModule
	}
	return nil
}

// versions records the version bumped from and the version bumped to.
func (r *bumpReport) versions(from, to string) {
	if r == nil {
		return
	}
	r.result.From, r.result.To = from, to
}

// tagged records tag, just created, and the commit it points at.
func (r *bumpReport) tagged(tag string) {
	if r == nil {
		return
	}
	r.result.Tag = tag
	r.result.Commit = localTagCommit(tag)
}

// pushed records that the tag reached origin: handlePushTag returned nil, which
// it does only once origin has the tag at the commit it points at here.
func (r *bumpReport) pushed() {
	if r == nil {
		return
	}
	r.result.Pushed = true
}

// finish prints the outcome as one JSON line, once the run is over, its switch
// back to the starting branch included; err is the error the run returned. It
// also gives the CLI logger back.
func (r *bumpReport) finish(err error) {
	if r == nil {
		return
	}

	if r.recorder != nil {
		pkglog.SetDefault(r.logger)
		r.result.Warnings = append(r.result.Warnings, r.recorder.kept()...)
	}

	r.result.Branch = currentBranchName()
	if r.result.Branch != r.result.StartBranch {
		r.result.Warnings = append(r.result.Warnings, fmt.Sprintf("The bump started on %s but ended on %s",
			describeBranch(r.result.StartBranch), describeBranch(r.result.Branch)))
	}

	if err != nil {
		r.result.Error = stripANSI(err.Error())
	}

	var line bytes.Buffer
	encoder := json.NewEncoder(&line)
	encoder.SetEscapeHTML(false)
	if encodeErr := encoder.Encode(r.result); encodeErr != nil {
		// Not expected: every field is a string, a bool, or a list of strings
		utils.Error("Failed to encode the version bump outcome as JSON: %v", encodeErr)
		return
	}
	utils.Print("%s", line.String()) // Encode ends the line
}

// warningRecorder stands in for the CLI logger while a bump with json runs. It
// passes every call on to the logger it replaces, so the output is unchanged,
// and keeps the text of each warning that logger prints.
type warningRecorder struct {
	pkglog.CLILogger

	mu       sync.Mutex
	warnings []string
}

// Warn prints the warning, and keeps its text.
func (w *warningRecorder) Warn(format string, args ...any) {
	w.CLILogger.Warn(format, args...)
	if w.GetLevel() > pkglog.LevelWarn {
		return // not printed
	}

	text := plainWarning(fmt.Sprintf(format, args...))
	if text == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.warnings = append(w.warnings, text)
}

// kept returns the text of the warnings printed so far.
func (w *warningRecorder) kept() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return slices.Clone(w.warnings)
}

// plainWarning returns a warning as plain text: without ANSI codes, and
// without the indentation, emoji, or bullet magex prints before it.
func plainWarning(msg string) string {
	return strings.TrimSpace(strings.TrimLeftFunc(stripANSI(msg), isWarningDecoration))
}

// isWarningDecoration reports whether r is part of what magex prints before a
// warning's text: indentation, an emoji with the selectors that shape it, or a
// bullet.
func isWarningDecoration(r rune) bool {
	return unicode.IsSpace(r) || unicode.In(r, unicode.So, unicode.Mn, unicode.Cf) || r == '•'
}

// stripANSI removes ANSI escape sequences, such as colors, from s.
func stripANSI(s string) string {
	const esc = '\x1b'
	if !strings.ContainsRune(s, esc) {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] != esc:
			b.WriteByte(s[i])
		case i+1 < len(s) && s[i+1] == '[':
			// A control sequence: ESC [, then parameter and intermediate bytes,
			// then a final byte from @ to ~ (m for a color), which the loop skips
			i += 2
			for i < len(s) && (s[i] < '@' || s[i] > '~') {
				i++
			}
		default:
			i++ // any other escape is ESC and one byte
		}
	}
	return b.String()
}

// currentBranchName returns the branch the clone is on, or "" when HEAD is
// detached or git can't tell.
func currentBranchName() string {
	branch, err := getCurrentBranch()
	if err != nil {
		return ""
	}
	return branch
}

// describeBranch names a branch in a warning, or a detached HEAD for "".
func describeBranch(name string) string {
	if name == "" {
		return "a detached HEAD"
	}
	return fmt.Sprintf("branch '%s'", name)
}
