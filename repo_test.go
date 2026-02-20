package main

import (
	"strings"
	"testing"
)

func TestRepo_Checkout(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	assertNoError(t, r.Checkout())
	runner.assertHasCalls(t, "git checkout main")
}

func TestRepo_Checkout_ErrorWhenNoRef(t *testing.T) {
	r, runner := fakeRepo(".", "", false)
	assertError(t, r.Checkout())
	runner.assertNoCalls(t)
}

func TestRepo_Config_NoEntriesNoCalls(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	assertNoError(t, r.Config(nil))
	runner.assertNoCalls(t)
}

func TestRepo_Config_OneEntryOneCall(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	gitCfg := map[string]string{"commit.gpgsign": "true"}
	assertNoError(t, r.Config(gitCfg))
	runner.assertHasCalls(t, "git config commit.gpgsign true")
}

func TestRepo_Config_MultipleEntriesMultipleCalls(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	gitCfg := map[string]string{
		"commit.gpgsign": "true",
		"foo":            "bar",
	}
	assertNoError(t, r.Config(gitCfg))
	runner.assertHasCalls(
		t,
		"git config commit.gpgsign true",
		"git config foo bar",
	)
}

func TestRepo_FastForward(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	assertNoError(t, r.FastForward())
	runner.assertHasCalls(t, "git pull --ff-only")
}

func TestRepo_Fetch(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	assertNoError(t, r.Fetch())
	runner.assertHasCalls(t, "git fetch -pt")
}

func TestRepo_Run_NoArgs(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	err := r.Run(nil)
	assertError(t, err)
	runner.assertNoCalls(t)
}

func TestRepo_Run_WithArgs(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	assertNoError(t, r.Run([]string{"echo", "hello"}))
	runner.assertHasCalls(t, "echo hello")
}

func TestRepo_Status(t *testing.T) {
	r, runner := fakeRepo(".", "main", false)
	assertNoError(t, r.Status())
	runner.assertHasCalls(t, "git status")
}

func TestRepo_UpdateSubmodules(t *testing.T) {
	r, runner := fakeRepo(".", "main", true)
	assertNoError(t, r.UpdateSubmodules())
	runner.assertHasCalls(t, "git submodule update --init --recursive")
}

func TestRepo_UpdateSubmodules_NoopWhenFalse(t *testing.T) {
	r, _ := fakeRepo(".", "main", false)
	assertNoError(t, r.UpdateSubmodules())
}

// Helpers

type fakeRunner struct {
	calls []string
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return nil
}

func (f *fakeRunner) assertHasCalls(t *testing.T, calls ...string) {
	expectedCount := len(calls)
	actualCount := len(f.calls)
	if expectedCount != actualCount {
		t.Errorf("expected %d call(s), actual %d: %#v", expectedCount, actualCount, f.calls)
	}
	for i, expectedCall := range calls {
		actualCall := f.calls[i]
		if expectedCall != actualCall {
			t.Errorf("expected call: %#v, actual call: %#v", expectedCall, actualCall)
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

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected error, got %#v", err)
	}
}

func fakeRepo(path string, ref string, submodules bool) (*Repo, *fakeRunner) {
	runner := &fakeRunner{}
	r := NewRepo(&RepoConfig{Path: path, Ref: ref, Submodules: submodules}, runner)
	return r, runner
}
