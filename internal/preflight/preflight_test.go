package preflight

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestAllPrerequisitesPassed(t *testing.T) {
	var lookedUp []string
	r := run("linux", "amd64", probes{
		openKVM: func(path string) error {
			if path != "/dev/kvm" {
				t.Errorf("checked KVM at %q", path)
			}
			return nil
		},
		findTool: func(name string) (string, error) {
			lookedUp = append(lookedUp, name)
			return "in ./bin", nil
		},
	})
	if r.OS != "linux" || r.Arch != "amd64" || !r.Ready() {
		t.Fatalf("unexpected report: %+v", r)
	}
	if want := []string{"adb", "launch_cvd", "stop_cvd"}; !reflect.DeepEqual(lookedUp, want) {
		t.Fatalf("tool probes = %v, want %v", lookedUp, want)
	}
	ids := []string{"platform", "kvm", "adb", "launch_cvd", "stop_cvd"}
	if len(r.Checks) != len(ids) {
		t.Fatalf("got %d checks, want %d", len(r.Checks), len(ids))
	}
	for i, check := range r.Checks {
		if check.ID != ids[i] || check.Status != Pass || check.Detail == "" {
			t.Errorf("check %d: %+v", i, check)
		}
	}
}

func TestWrongPlatformSkipsHostProbes(t *testing.T) {
	for _, platform := range [][2]string{{"windows", "amd64"}, {"darwin", "arm64"}, {"linux", "arm64"}} {
		t.Run(platform[0]+"-"+platform[1], func(t *testing.T) {
			r := run(platform[0], platform[1], probes{
				openKVM: func(string) error { t.Fatal("unexpected KVM probe"); return nil },
				findTool: func(string) (string, error) {
					t.Fatal("unexpected tool probe")
					return "", nil
				},
			})
			if r.Ready() || len(r.Checks) != 5 || r.Checks[0].Status != Fail {
				t.Fatalf("unexpected report: %+v", r)
			}
			for _, check := range r.Checks[1:] {
				if check.Status != Skip {
					t.Errorf("%s = %s, want skip", check.ID, check.Status)
				}
			}
		})
	}
}

func TestKVMFailuresAndToolFailuresAreIndependent(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want string
	}{
		{"missing", os.ErrNotExist, "not available"},
		{"permission", os.ErrPermission, "cannot be opened read-write"},
		{"wrong type", errNotCharDevice, "not a character device"},
		{"unexpected", errors.New("I/O failure"), "I/O failure"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := run("linux", "amd64", probes{
				openKVM: func(path string) error {
					return &os.PathError{Op: "open", Path: path, Err: tt.err}
				},
				findTool: func(name string) (string, error) {
					if name == "launch_cvd" {
						return "", os.ErrNotExist
					}
					return "on PATH", nil
				},
			})
			if r.Ready() {
				t.Fatal("failed prerequisites cannot report ready")
			}
			if k := r.Checks[1]; k.Status != Fail || !strings.Contains(k.Detail, tt.want) {
				t.Errorf("KVM check = %+v, want failure containing %q", k, tt.want)
			}
			if r.Checks[2].Status != Pass || r.Checks[3].Status != Fail || r.Checks[4].Status != Pass {
				t.Fatalf("tool checks = %+v", r.Checks[2:])
			}
		})
	}
}

func TestKVMProbeDoesNotAcceptRegularOrMissingFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "kvm")
	if err := checkKVM(file); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing: %v", err)
	}
	if err := os.WriteFile(file, []byte("not a device"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkKVM(file); !errors.Is(err, errNotCharDevice) {
		t.Fatalf("regular file: %v", err)
	}
	if got, err := os.ReadFile(file); err != nil || string(got) != "not a device" {
		t.Fatalf("probe modified the file: %q, %v", got, err)
	}
}

func TestFindToolDoesNotRunExecutables(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the M0 host and its executable mode checks are Linux-specific")
	}
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	name := "android4android-preflight-fixture"
	tool := filepath.Join(dir, name)
	// Deliberately not a valid executable image.
	if err := os.WriteFile(tool, []byte("cannot be executed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := findTool(name); err == nil {
		t.Fatal("accepted a file without execute bits")
	}
	if err := os.Chmod(tool, 0700); err != nil {
		t.Fatal(err)
	}
	if got, err := findTool(name); err != nil || got != "on PATH" {
		t.Fatalf("findTool = %q, %v", got, err)
	}
	local := t.TempDir()
	if err := os.WriteFile(filepath.Join(local, name), []byte("not an executable image"), 0700); err != nil {
		t.Fatal(err)
	}
	if got, err := findToolIn(local, name); err != nil || got != "in ./bin" {
		t.Fatalf("local package should take precedence, got %q, %v", got, err)
	}
}

func TestReportJSONAndReadiness(t *testing.T) {
	if (Report{}).Ready() || (Report{Checks: []Result{{Status: Skip}}}).Ready() {
		t.Fatal("empty and skipped reports cannot be ready")
	}
	r := Report{OS: "linux", Arch: "amd64", Checks: []Result{{ID: "platform", Status: Pass, Detail: "test"}}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"os":"linux","arch":"amd64","checks":[{"id":"platform","status":"pass","detail":"test"}]}`; string(data) != want {
		t.Errorf("JSON = %s, want %s", data, want)
	}
}
