package landlock

import (
	"testing"

	ll "github.com/landlock-lsm/go-landlock/landlock/syscall"
)

func TestConfigString(t *testing.T) {
	for _, tc := range []struct {
		cfg  Config
		want string
	}{
		{
			cfg:  Config{handledAccessFS: 0, handledAccessNet: 0},
			want: "{Landlock V0; FS: ∅; Net: ∅; Scope: ∅}",
		},
		{
			cfg:  Config{handledAccessFS: ll.AccessFSWriteFile},
			want: "{Landlock V1; FS: {write_file}; Net: ∅; Scope: ∅}",
		},
		{
			cfg:  Config{handledAccessNet: ll.AccessNetBindTCP},
			want: "{Landlock V4; FS: ∅; Net: {bind_tcp}; Scope: ∅}",
		},
		{
			cfg:  V1,
			want: "{Landlock V1; FS: all; Net: ∅; Scope: ∅}",
		},
		{
			cfg:  V1.BestEffort(),
			want: "{Landlock V1; FS: all; Net: ∅; Scope: ∅ (best effort)}",
		},
		{
			cfg:  Config{handledScopes: ScopeSignal},
			want: "{Landlock V6; FS: ∅; Net: ∅; Scope: {signal}}",
		},
		{
			cfg:  Config{handledAccessFS: 1 << 63},
			want: "{Landlock V???; FS: {1<<63}; Net: ∅; Scope: ∅}",
		},
	} {
		got := tc.cfg.String()
		if got != tc.want {
			t.Errorf("cfg.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestNewConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []interface{}
		want Config
	}{
		{
			name: "AccessFS",
			args: []interface{}{AccessFSSet(ll.AccessFSWriteFile)},
			want: Config{handledAccessFS: ll.AccessFSWriteFile},
		},
		{
			name: "AccessNet",
			args: []interface{}{AccessNetSet(ll.AccessNetBindTCP)},
			want: Config{handledAccessNet: ll.AccessNetBindTCP},
		},
		{
			name: "Scope",
			args: []interface{}{ScopeSignal},
			want: Config{handledScopes: ScopeSignal},
		},
		{
			name: "Mixed",
			args: []interface{}{AccessFSSet(ll.AccessFSWriteFile), ScopeSignal},
			want: Config{handledAccessFS: ll.AccessFSWriteFile, handledScopes: ScopeSignal},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewConfig(tc.args...)
			if err != nil {
				t.Fatalf("NewConfig(): expected success, got %v", err)
			}
			if c.handledAccessFS != tc.want.handledAccessFS {
				t.Errorf("handledAccessFS = %v, want %v", c.handledAccessFS, tc.want.handledAccessFS)
			}
			if c.handledAccessNet != tc.want.handledAccessNet {
				t.Errorf("handledAccessNet = %v, want %v", c.handledAccessNet, tc.want.handledAccessNet)
			}
			if c.handledScopes != tc.want.handledScopes {
				t.Errorf("handledScopes = %v, want %v", c.handledScopes, tc.want.handledScopes)
			}
		})
	}
}

func TestNewConfigEmpty(t *testing.T) {
	// Constructing an empty config is a bit pointless, but should work.
	c, err := NewConfig()
	if err != nil {
		t.Errorf("NewConfig(): expected success, got %v", err)
	}
	if c.handledAccessFS != 0 {
		t.Errorf("c.handledAccessFS = %v, want 0", c.handledAccessFS)
	}
	if c.handledAccessNet != 0 {
		t.Errorf("c.handledAccessNet = %v, want 0", c.handledAccessNet)
	}
	if c.handledScopes != 0 {
		t.Errorf("c.handledScopes = %v, want 0", c.handledScopes)
	}
}

func TestNewConfigFailures(t *testing.T) {
	for _, args := range [][]interface{}{
		{ll.AccessFSWriteFile},
		{123},
		{"a string"},
		{"foo", 42},
		// May not specify two AccessFSSets
		{AccessFSSet(ll.AccessFSWriteFile), AccessFSSet(ll.AccessFSReadFile)},
		// May not specify an unsupported AccessFSSet value
		{AccessFSSet(1 << 16)},
		{AccessFSSet(1 << 63)},
		// May not specify two ScopeSets
		{ScopeSignal, ScopeAbstractUnixSocket},
		// May not specify an unsupported ScopeSet value
		{ScopeSet(1 << 5)},
	} {
		_, err := NewConfig(args...)
		if err == nil {
			t.Errorf("NewConfig(%v) success, expected error", args)
		}
	}
}
