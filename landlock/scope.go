package landlock

import ll "github.com/landlock-lsm/go-landlock/landlock/syscall"

var scopeNames = []string{
	"abstract_unix_socket",
	"signal",
}

// ScopeSet is a set of Landlock scopes.
type ScopeSet uint64

// Scope constants that can be combined to form a ScopeSet.
const (
	ScopeAbstractUnixSocket ScopeSet = ScopeSet(ll.ScopeAbstractUnixSocket)
	ScopeSignal             ScopeSet = ScopeSet(ll.ScopeSignal)
)

var supportedScopes = ScopeSet(ScopeAbstractUnixSocket | ScopeSignal)

func (s ScopeSet) String() string {
	return accessSetString(uint64(s), scopeNames)
}

func (s ScopeSet) isSubset(other ScopeSet) bool {
	return s&other == s
}

func (s ScopeSet) intersect(other ScopeSet) ScopeSet {
	return s & other
}

func (s ScopeSet) union(other ScopeSet) ScopeSet {
	return s | other
}

func (s ScopeSet) isEmpty() bool {
	return s == 0
}

// valid returns true iff the given ScopeSet is supported by this version of go-landlock.
func (s ScopeSet) valid() bool {
	return s.isSubset(supportedScopes)
}
