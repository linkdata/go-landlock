//go:build linux

package landlock_test

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/landlock-lsm/go-landlock/landlock"
	"github.com/landlock-lsm/go-landlock/landlock/lltest"
	"golang.org/x/sys/unix"
)

func TestScopeSignal(t *testing.T) {
	lltest.RunInSubprocess(t, func() {
		lltest.RequireABI(t, 6)

		target := os.Getppid()
		if err := unix.Kill(target, 0); err != nil {
			t.Skipf("kill(0) before Landlock failed: %v", err)
		}

		cfg := landlock.MustConfig(landlock.ScopeSignal)
		if err := cfg.Restrict(); err != nil {
			t.Fatalf("landlock.Restrict: %v", err)
		}

		if err := unix.Kill(target, 0); !errors.Is(err, unix.EPERM) {
			t.Fatalf("kill(0) after Landlock: got %v, want EPERM", err)
		}
	})
}

func TestScopeAbstractUnixSocket(t *testing.T) {
	lltest.RunInSubprocess(t, func() {
		lltest.RequireABI(t, 6)

		socketName := fmt.Sprintf("@go-landlock-%d", time.Now().UnixNano())

		cmd := exec.Command(os.Args[0], "-test.run=TestScopeAbstractUnixSocketHelper$")
		cmd.Env = append(os.Environ(), "LANDLOCK_ABSTRACT_SOCKET="+socketName)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatalf("StdoutPipe: %v", err)
		}
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			t.Fatalf("start helper: %v", err)
		}
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		reader := bufio.NewReader(stdout)
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("waiting for helper: %v", err)
		}
		if strings.TrimSpace(line) != "READY" {
			t.Fatalf("unexpected helper output %q", line)
		}

		if err := dialAbstractUnix(socketName); err != nil {
			t.Fatalf("dial before Landlock: %v", err)
		}

		cfg := landlock.MustConfig(landlock.ScopeAbstractUnixSocket)
		if err := cfg.Restrict(); err != nil {
			t.Fatalf("landlock.Restrict: %v", err)
		}

		if err := dialAbstractUnix(socketName); !errors.Is(err, unix.EPERM) {
			t.Fatalf("dial after Landlock: got %v, want EPERM", err)
		}
	})
}

func TestScopeAbstractUnixSocketHelper(t *testing.T) {
	name := os.Getenv("LANDLOCK_ABSTRACT_SOCKET")
	if name == "" {
		t.Skip("helper process")
	}

	addr := &net.UnixAddr{Name: name, Net: "unix"}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		t.Fatalf("ListenUnix: %v", err)
	}
	defer listener.Close()

	fmt.Println("READY")

	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return
		}
		conn.Close()
	}
}

func dialAbstractUnix(name string) error {
	addr := &net.UnixAddr{Name: name, Net: "unix"}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
