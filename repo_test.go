package main

import (
	"slices"
	"strings"
	"testing"
)

func TestRepo_Checkout(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.Checkout())
	runner.assertHasCalls(t, "git checkout main")
}

func TestRepo_Checkout_ErrorWhenNoRef(t *testing.T) {
	r, runner := fakeRepo(".", "git@github.com:Foo/bar.git", "")
	assertError(t, r.Checkout())
	runner.assertNoCalls(t)
}

func TestRepo_Clone_CloneWhenDirDoesNotExist(t *testing.T) {
	r, runner := fakeRepo("/fake/dir/plz", "git@github.com:Foo/bar.git", "develop")
	r.Clone()
	runner.assertHasCalls(t, "git clone git@github.com:Foo/bar.git /fake/dir/plz")
}

func TestRepo_Clone_NoCloneWhenDirExists(t *testing.T) {
	tmp := t.TempDir()
	r, runner := fakeRepo(tmp, "git@github.com:Foo/bar.git", "develop")
	r.Clone()
	runner.assertNoCalls(t)
}

func TestRepo_Config_NoEntriesNoCalls(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.Config(nil))
	runner.assertNoCalls(t)
}

func TestRepo_Config_OneEntryOneCall(t *testing.T) {
	r, runner := defaultFakeRepo()
	gitCfg := map[string]string{"commit.gpgsign": "true"}
	assertNotError(t, r.Config(gitCfg))
	runner.assertHasCalls(t, "git config commit.gpgsign true")
}

func TestRepo_Config_MultipleEntriesMultipleCalls(t *testing.T) {
	r, runner := defaultFakeRepo()
	gitCfg := map[string]string{
		"commit.gpgsign": "true",
		"foo":            "bar",
	}
	assertNotError(t, r.Config(gitCfg))
	runner.assertHasCallsUnordered(
		t,
		"git config commit.gpgsign true",
		"git config foo bar",
	)
}

func TestRepo_FastForward(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.FastForward())
	runner.assertHasCalls(t, "git pull --ff-only")
}

func TestRepo_Fetch(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.Fetch())
	runner.assertHasCalls(t, "git fetch -pt")
}

func TestRepo_Run_NoArgs(t *testing.T) {
	r, runner := defaultFakeRepo()
	err := r.Run(nil)
	assertError(t, err)
	runner.assertNoCalls(t)
}

func TestRepo_Run_WithArgs(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.Run([]string{"echo", "hello"}))
	runner.assertHasCalls(t, "echo hello")
}

func TestRepo_Status(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.Status())
	runner.assertHasCalls(t, "git status")
}

func TestRepo_UpdateSubmodules(t *testing.T) {
	r, runner := defaultFakeRepo()
	assertNotError(t, r.UpdateSubmodules())
	runner.assertHasCalls(t, "git submodule update --init --recursive")
}

// Fake runner

type fakeRunner struct {
	calls []string
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return nil
}

// Assertions

func (f *fakeRunner) assertCallCount(t *testing.T, calls ...string) {
	expectedCount := len(calls)
	actualCount := len(f.calls)
	if expectedCount != actualCount {
		t.Errorf("expected %d call(s), actual %d: %#v", expectedCount, actualCount, f.calls)
	}
}

func (f *fakeRunner) assertHasCalls(t *testing.T, calls ...string) {
	f.assertCallCount(t, calls...)
	for i, expectedCall := range calls {
		actualCall := f.calls[i]
		if expectedCall != actualCall {
			t.Errorf("expected call: %#v, actual call: %#v", expectedCall, actualCall)
		}
	}
}

func (f *fakeRunner) assertHasCallsUnordered(t *testing.T, calls ...string) {
	expectedCount := len(calls)
	actualCount := len(f.calls)
	if expectedCount != actualCount {
		t.Errorf("expected %d call(s), actual %d: %#v", expectedCount, actualCount, f.calls)
	}
	for _, expectedCall := range calls {
		if !slices.Contains(f.calls, expectedCall) {
			t.Errorf("expected call: %#v, actual calls: %#v", expectedCall, f.calls)
		}
	}
}

func (f *fakeRunner) assertNoCalls(t *testing.T) {
	callCount := len(f.calls)
	if callCount > 0 {
		t.Errorf("expected no calls, actual %d: %#v", callCount, f.calls)
	}
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func assertNotError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected nil, got %#v", err)
	}
}

// Create fake repos

func defaultFakeRepo() (*Repo, *fakeRunner) {
	return fakeRepo(".", "git@github.com:Foo/bar.git", "main")
}

func fakeRepo(path string, remote string, ref string) (*Repo, *fakeRunner) {
	runner := &fakeRunner{}
	r := NewRepo(RepoConfig{Path: path, Remote: remote, Ref: ref}, runner)
	return r, runner
}
