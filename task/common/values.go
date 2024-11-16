package common

import (
	"errors"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gobwas/glob"
)

var NotFoundError error = errors.New("resource not found")
var NotImplementedError error = errors.New("resource method not implemented")

// <0=disabled, 0=auto, >0=fixed
type Spot float64

const (
	SpotDisabled Spot = -1
	SpotEnabled  Spot = 0
)

type Status map[StatusCode]int

type StatusCode string

const (
	StatusCodeActive    StatusCode = "running"
	StatusCodeSucceeded StatusCode = "succeeded"
	StatusCodeFailed    StatusCode = "failed"
	StatusCodeTimeout   StatusCode = "timeout"
)

type Size struct {
	Storage int
	Machine string
}

type Event struct {
	Time        time.Time
	Code        string
	Description []string
}

// RemoteStorage contains the configuration for the cloud storage container
// used by the task.
type RemoteStorage struct {
	// Container stores the id of the container to be used.
	Container string
	// Path stores the subdirectory inside the container.
	Path string
	// Config stores provider-specific configuration keys for accessing the pre-allocated
	// storage container.
	Config map[string]string
}

type Task struct {
	Size          Size
	Environment   Environment
	Firewall      Firewall
	PermissionSet string
	Spot          Spot
	Parallelism   uint16

	RemoteStorage *RemoteStorage

	Addresses []net.IP
	Status    Status
	Events    []Event
}

// Firewall
type Firewall struct {
	Ingress FirewallRule
	Egress  FirewallRule
}

// Firewall rule: not specified fields mean "allow any"; sepcified but empty mean "allow none";
// ports are both TCP and UDP; when ports is not specified, it will allow ingress to every port
// and every protocol, not only TCP&UDP
type FirewallRule struct {
	Nets  *[]net.IPNet
	Ports *[]uint16
}

type Environment struct {
	Image        string
	Script       string
	Variables    Variables
	Timeout      time.Duration
	Directory    string
	DirectoryOut string
	ExcludeList  []string
}

type Variables map[string]*string

// Enrich takes a map[string]*string of environment variables and, when a map value
// is <nil>, tries to get the value from the process environment variables. If
// the map key is a valid glob and the value is <nil>, all the matching environment
// variables will be set in the resulting map.
func (v Variables) Enrich() map[string]string {
	result := make(map[string]string)
	for name, value := range map[string]*string(v) {
		if value == nil {
			// FIXME: remove Replace and QuoteMeta to enable extended glob.
			g := glob.MustCompile(strings.ReplaceAll(glob.QuoteMeta(name), `\*`, `*`))
			for _, variable := range os.Environ() {
				if key := strings.Split(variable, "=")[0]; g.Match(key) {
					result[key] = os.Getenv(key)
				}
			}
		} else {
			result[name] = *value
		}
	}
	return result
}
