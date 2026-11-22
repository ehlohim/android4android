// Package preflight checks local prerequisites for the first Cuttlefish smoke test.
package preflight

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Status describes the outcome of one prerequisite check.
type Status string

const (
	Pass Status = "pass"
	Fail Status = "fail"
	Skip Status = "skip"
)

// Result describes a single prerequisite and the reason for its status.
type Result struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	Detail string `json:"detail"`
}

// Report describes the current host. It does not verify that a guest can
// actually boot, that an image is installed, or that audio and SIP work.
type Report struct {
	OS     string   `json:"os"`
	Arch   string   `json:"arch"`
	Checks []Result `json:"checks"`
}

// Ready reports whether all checks in this nonempty report passed.
func (r Report) Ready() bool {
	if len(r.Checks) == 0 {
		return false
	}
	for _, check := range r.Checks {
		if check.Status != Pass {
			return false
		}
	}
	return true
}

type probes struct {
	openKVM  func(string) error
	findTool func(string) (string, error)
}

var tools = [...]string{"adb", "launch_cvd", "stop_cvd"}

// Check inspects the current host without starting processes or writing to
// KVM. The first smoke test targets Linux on x86-64 (Go's amd64), regardless
// of the platforms that Cuttlefish or this project may support later.
func Check() Report {
	return run(runtime.GOOS, runtime.GOARCH, probes{checkKVM, findTool})
}

func run(goos, goarch string, p probes) Report {
	r := Report{OS: goos, Arch: goarch, Checks: make([]Result, 0, 2+len(tools))}
	if goos != "linux" || goarch != "amd64" {
		r.Checks = append(r.Checks, Result{"platform", Fail, "The M0 Cuttlefish host must be Linux on x86-64 (amd64)."})
		r.Checks = append(r.Checks, Result{"kvm", Skip, "Not checked on a platform outside the M0 target."})
		for _, name := range tools {
			r.Checks = append(r.Checks, Result{name, Skip, "Not checked on a platform outside the M0 target."})
		}
		return r
	}

	r.Checks = append(r.Checks, Result{"platform", Pass, "Linux on x86-64 (amd64)."})
	r.Checks = append(r.Checks, kvmResult(p.openKVM("/dev/kvm")))
	for _, name := range tools {
		location, err := p.findTool(name)
		if err != nil {
			r.Checks = append(r.Checks, Result{name, Fail, fmt.Sprintf("%s was not found as an executable in ./bin or on PATH. Run from the extracted Cuttlefish host package directory or add its bin directory to PATH.", name)})
			continue
		}
		r.Checks = append(r.Checks, Result{name, Pass, fmt.Sprintf("%s is present %s; it was not run.", name, location)})
	}
	return r
}

var errNotCharDevice = errors.New("not a character device")

func kvmResult(err error) Result {
	switch {
	case err == nil:
		return Result{"kvm", Pass, "/dev/kvm opened read-write as the current user; no VM was created."}
	case errors.Is(err, os.ErrNotExist):
		return Result{"kvm", Fail, "/dev/kvm is not available. The host or container must expose KVM."}
	case errors.Is(err, os.ErrPermission):
		return Result{"kvm", Fail, "/dev/kvm cannot be opened read-write as the current user. Check host permissions or device policy."}
	case errors.Is(err, errNotCharDevice):
		return Result{"kvm", Fail, "/dev/kvm is not a character device."}
	default:
		return Result{"kvm", Fail, fmt.Sprintf("Cannot check read-write access to /dev/kvm: %v", err)}
	}
}

func checkKVM(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return errNotCharDevice
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return errNotCharDevice
	}
	return nil
}

func findTool(name string) (string, error) {
	return findToolIn("bin", name)
}

func findToolIn(binDir, name string) (string, error) {
	// Cuttlefish's official host archive keeps its executables in ./bin.
	// Use an explicit path so LookPath does not search PATH for this check.
	local := filepath.Join(binDir, name)
	if _, err := exec.LookPath(local); err == nil {
		return "in ./bin", nil
	}
	if path, err := exec.LookPath(name); err == nil && filepath.IsAbs(path) {
		return "on PATH", nil
	}
	return "", exec.ErrNotFound
}
