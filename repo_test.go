package main

import (
	"slices"
	"strings"
	"testing"
)

func TestRepoCommands(t *testing.T) {
	tests := []struct {
		name  string
		cmd   func(Repo) error
		calls []string
	}{
		{"Checkout", Repo.Checkout, []string{"git checkout main"}},
		{"FastForward", Repo.FastForward, []string{"git pull --ff-only"}},
		{"Fetch", Repo.Fetch, []string{"git fetch"}},
		{"Push", Repo.Push, []string{"git push"}},
		{"Status", Repo.Status, []string{"git status"}},
		{
			"UpdateSubmodules",
			Repo.UpdateSubmodules,
			[]string{"git submodule update --init --recursive"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, runner := defaultFakeRepo()
			if err := tt.cmd(r); err != nil {
				t.Errorf("err=%v", err)
			}
			runner.AssertHasCalls(t, tt.calls...)
		})
	}
}

func TestRepoCheckoutErrorWhenNoRef(t *testing.T) {
	r, runner := fakeRepo(".", "git@github.com:Foo/bar.git", "")
	if err := r.Checkout(); err == nil {
		t.Errorf("err=nil")
	}
	runner.AssertNoCalls(t)
}

func TestRepoCloneCloneWhenDirDoesNotExist(t *testing.T) {
	r, runner := fakeRepo("/fake/dir/plz", "git@github.com:Foo/bar.git", "develop")
	if err := r.Clone(); err != nil {
		t.Errorf("err=%v", err)
	}
	runner.AssertHasCalls(t, "git clone git@github.com:Foo/bar.git /fake/dir/plz")
}

func TestRepoCloneNoCloneWhenDirExists(t *testing.T) {
	tmp := t.TempDir()
	r, runner := fakeRepo(tmp, "git@github.com:Foo/bar.git", "develop")
	if err := r.Clone(); err != nil {
		t.Errorf("err=%v", err)
	}
	runner.AssertNoCalls(t)
}

func TestRepoConfigNoEntriesNoCalls(t *testing.T) {
	r, runner := defaultFakeRepo()
	if err := r.Config(nil); err != nil {
		t.Errorf("err=%v", err)
	}
	runner.AssertNoCalls(t)
}

func TestRepoConfigOneEntryOneCall(t *testing.T) {
	r, runner := defaultFakeRepo()
	gitCfg := map[string]string{"commit.gpgsign": "true"}
	if err := r.Config(gitCfg); err != nil {
		t.Errorf("err=%v", err)
	}
	runner.AssertHasCalls(t, "git config commit.gpgsign true")
}

func TestRepoConfigMultipleEntriesMultipleCalls(t *testing.T) {
	r, runner := defaultFakeRepo()
	gitCfg := map[string]string{
		"commit.gpgsign": "true",
		"foo":            "bar",
	}
	if err := r.Config(gitCfg); err != nil {
		t.Errorf("err=%v", err)
	}
	runner.AssertHasCallsUnordered(
		t,
		"git config commit.gpgsign true",
		"git config foo bar",
	)
}

func TestRepoRunNoArgs(t *testing.T) {
	r, runner := defaultFakeRepo()
	if err := r.Run(nil); err == nil {
		t.Errorf("err=nil")
	}
	runner.AssertNoCalls(t)
}

func TestRepoRunWithArgs(t *testing.T) {
	r, runner := defaultFakeRepo()
	if err := r.Run([]string{"echo", "hello"}); err != nil {
		t.Errorf("err=%v", err)
	}
	runner.AssertHasCalls(t, "echo hello")
}

type FakeRunner struct {
	calls []string
}

func (f *FakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return nil
}

func (f *FakeRunner) AssertHasCalls(t *testing.T, calls ...string) {
	t.Helper()
	f.assertCallCount(t, calls)
	for i, expectedCall := range calls {
		actualCall := f.calls[i]
		if expectedCall != actualCall {
			t.Errorf("expected call: %#v, actual call: %#v", expectedCall, actualCall)
		}
	}
}

func (f *FakeRunner) AssertHasCallsUnordered(t *testing.T, calls ...string) {
	t.Helper()
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

func (f *FakeRunner) AssertNoCalls(t *testing.T) {
	t.Helper()
	f.assertCallCount(t, []string{})
}

func (f *FakeRunner) assertCallCount(t *testing.T, calls []string) {
	t.Helper()
	expectedCount := len(calls)
	actualCount := len(f.calls)
	if expectedCount != actualCount {
		t.Fatalf("expected %d call(s), actual %d: %#v", expectedCount, actualCount, f.calls)
	}
}

// Create fake repos

func defaultFakeRepo() (Repo, *FakeRunner) {
	return fakeRepo(".", "git@github.com:Foo/bar.git", "main")
}

func fakeRepo(path string, remote string, ref string) (Repo, *FakeRunner) {
	runner := &FakeRunner{}
	r := Repo{RepoConfig{path, remote, ref}, runner}
	return r, runner
}
