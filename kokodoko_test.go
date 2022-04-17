package kokodoko_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kinbiko/bugsnag"

	"github.com/kinbiko/kokodoko"
)

func TestKokodoko(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := kokodoko.Config{}
	n, _ := bugsnag.New(bugsnag.Configuration{
		APIKey:                 "1234abcd1234abcd1234abcd1234abcd",
		AppVersion:             "0.0.1",
		ReleaseStage:           "Test",
		ErrorReportSanitizer:   func(context.Context, *bugsnag.JSONErrorReport) error { return fmt.Errorf("don't send") },
		SessionReportSanitizer: func(*bugsnag.JSONSessionReport) error { return fmt.Errorf("don't send") },
	})

	t.Run("success case", func(t *testing.T) {
		var (
			dirArg  = "./cmd/kokodoko/"
			pathArg = "./cmd/kokodoko/main.go"
		)
		mockSystem := &SystemMock{
			RemoteURLFunc: func(context.Context, string) (string, error) { return "https://github.com/kinbiko/kokodoko", nil },
			HashFunc:      func(context.Context, string) (string, error) { return "565983f8815aa3919bfc219dca7b692d0509911f", nil },
			RepoRootFunc:  func(context.Context, string) (string, error) { return "/Users/roger/repos/kokodoko", nil },
			AbsolutePathFunc: func(s string) (string, error) {
				base := "/Users/roger/repos/kokodoko/cmd/kokodoko"
				if s == pathArg {
					return base + "/main.go", nil
				}
				return base, nil
			},
		}

		var (
			expDirURL = "https://github.com/kinbiko/kokodoko/blob/565983f8815aa3919bfc219dca7b692d0509911f/cmd/kokodoko"
			expFile   = "/main.go"
		)
		for _, tc := range []struct { // nolint:govet // Don't care about perf. It's just easier to read like this
			name string
			args []string
			exp  string
		}{
			{"only dir", []string{dirArg}, expDirURL},
			{"only file", []string{pathArg}, expDirURL + expFile},
			{"path with one line", []string{pathArg, "12"}, expDirURL + expFile + "#L12"},
			{"path with line range", []string{pathArg, "12-30"}, expDirURL + expFile + "#L12-L30"},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, tc.args)
				if err != nil {
					t.Error(err)
				}
				if got != tc.exp {
					t.Errorf("expected url:\n%s\nbut got:\n%s", tc.exp, got)
				}
			})
		}
	})

	t.Run("failures", func(t *testing.T) {
		t.Parallel()

		t.Run("argument error", func(t *testing.T) {
			for _, tc := range []struct {
				name string
				args []string
			}{
				{"no args", nil},
				{"empty args", []string{}},
				{"too many args", []string{"foo", "bar", "baz"}},
			} {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					mockSystem := &SystemMock{}
					got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, tc.args)
					if err == nil {
						t.Errorf("expected error but got: %s", got)
					}
				})
			}
		})

		t.Run("file doesn't exist", func(t *testing.T) {
			t.Parallel()

			mockSystem := &SystemMock{}
			got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, []string{"filepath that most certainly does not exist yo"})
			if err == nil {
				t.Errorf("expected error but got: %s", got)
			}
		})

		t.Run("directory with line numbers", func(t *testing.T) {
			t.Parallel()

			mockSystem := &SystemMock{}
			got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, []string{"./cmd/", "16"})
			if err == nil {
				t.Errorf("expected error but got: %s", got)
			}
		})

		t.Run("error on requesting remote URL", func(t *testing.T) {
			t.Parallel()

			mockSystem := &SystemMock{
				RemoteURLFunc: func(context.Context, string) (string, error) { return "", fmt.Errorf("something went wrong") },
				HashFunc:      func(context.Context, string) (string, error) { return "565983f8815aa3919bfc219dca7b692d0509911f", nil },
				RepoRootFunc:  func(context.Context, string) (string, error) { return "/Users/roger/repos/kokodoko", nil },
			}
			got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, []string{"./cmd/kokodoko/main.go", "16"})
			if err == nil {
				t.Errorf("expected error but got: %s", got)
			}
		})

		t.Run("error on requesting hash", func(t *testing.T) {
			t.Parallel()

			mockSystem := &SystemMock{
				RemoteURLFunc: func(context.Context, string) (string, error) { return "https://github.com/kinbiko/kokodoko", nil },
				HashFunc:      func(context.Context, string) (string, error) { return "", fmt.Errorf("something went wrong") },
				RepoRootFunc:  func(context.Context, string) (string, error) { return "/Users/roger/repos/kokodoko", nil },
			}
			got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, []string{"./cmd/kokodoko/main.go", "16"})
			if err == nil {
				t.Errorf("expected error but got: %s", got)
			}
		})

		t.Run("error on requesting repo root", func(t *testing.T) {
			t.Parallel()

			mockSystem := &SystemMock{
				RemoteURLFunc: func(context.Context, string) (string, error) { return "https://github.com/kinbiko/kokodoko", nil },
				HashFunc:      func(context.Context, string) (string, error) { return "565983f8815aa3919bfc219dca7b692d0509911f", nil },
				RepoRootFunc:  func(context.Context, string) (string, error) { return "", fmt.Errorf("something went wrong") },
			}
			got, err := kokodoko.New(mockSystem, n, cfg).Run(ctx, []string{"./cmd/kokodoko/main.go", "16"})
			if err == nil {
				t.Errorf("expected error but got: %s", got)
			}
		})
	})
}

type SystemMock struct {
	RemoteURLFunc    func(ctx context.Context, repoPath string) (string, error)
	HashFunc         func(ctx context.Context, repoPath string) (string, error)
	RepoRootFunc     func(ctx context.Context, repoPath string) (string, error)
	AbsolutePathFunc func(relative string) (string, error)
}

func (m *SystemMock) RemoteURL(ctx context.Context, repoPath string) (string, error) {
	if m.RemoteURLFunc == nil {
		panic("unexpected call to RemoteURL")
	}
	return m.RemoteURLFunc(ctx, repoPath)
}
func (m *SystemMock) Hash(ctx context.Context, repoPath string) (string, error) {
	if m.HashFunc == nil {
		panic("unexpected call to Hash")
	}
	return m.HashFunc(ctx, repoPath)
}
func (m *SystemMock) RepoRoot(ctx context.Context, repoPath string) (string, error) {
	if m.RepoRootFunc == nil {
		panic("unexpected call to RepoRoot")
	}
	return m.RepoRootFunc(ctx, repoPath)
}
func (m *SystemMock) AbsolutePath(relative string) (string, error) {
	if m.AbsolutePathFunc == nil {
		panic("unexpected call to AbsolutePath")
	}
	return m.AbsolutePathFunc(relative)
}
