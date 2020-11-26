package kokodoko_test

import (
	"context"
	"testing"

	"github.com/kinbiko/bugsnag"
	"github.com/kinbiko/kokodoko"
)

func TestKokodoko(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	t.Run("success case", func(t *testing.T) {
		mockSystem := &SystemMock{
			RemoteURLFunc: func(context.Context, string) (string, error) { return "https://github.com/kinbiko/kokodoko", nil },
			HashFunc:      func(context.Context, string) (string, error) { return "565983f8815aa3919bfc219dca7b692d0509911f", nil },
			RepoRootFunc:  func(context.Context, string) (string, error) { return "/Users/roger/repos/kokodoko", nil },
		}
		n, _ := bugsnag.New(bugsnag.Configuration{
			APIKey:       "1234abcd1234abcd1234abcd1234abcd",
			AppVersion:   "0.0.1",
			ReleaseStage: "Test",
		})
		app := kokodoko.New(mockSystem, n, kokodoko.Config{})
		pathArg := "./cmd/kokodoko/main.go"
		expURL := "https://github.com/kinbiko/kokodoko/blob/565983f8815aa3919bfc219dca7b692d0509911f/cmd/kokodoko/main.go"
		for _, tc := range []struct {
			name string
			args []string
			exp  string
		}{
			{"only path", []string{pathArg}, expURL},
			{"path with one line", []string{pathArg, "12"}, expURL + "#L12"},
			{"path with line range", []string{pathArg, "12-30"}, expURL + "#L12-L30"},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				got, err := app.Run(ctx, tc.args)
				if err != nil {
					t.Error(err)
				}
				if got != tc.exp {
					t.Errorf("expected url:\n%s\nbut got:\n%s", tc.exp, got)
				}
			})
		}
	})
}

type SystemMock struct {
	RemoteURLFunc func(ctx context.Context, repoPath string) (string, error)
	HashFunc      func(ctx context.Context, repoPath string) (string, error)
	RepoRootFunc  func(ctx context.Context, repoPath string) (string, error)
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
