package vnc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mitchellh/go-vnc"
	"github.com/s-newman/scorestack/dynamicbeat/checks/schema"
)

// The Definition configures the behavior of the VNC check
// it implements the "check" interface
type Definition struct {
	ID          string  // a unique identifier for this check
	Name        string  // a human-readable title for the check
	Group       string  // the group this check is part of
	ScoreWeight float64 // the weight that this check has relative to others
	Host        string  // (required) The IP or hostname of the vnc server
	Port        string  // (required) The port for the vnc server
	Password    string  // (required) The password for the vnc server
}

// Run a single instance of the check
func (d *Definition) Run(ctx context.Context, wg *sync.WaitGroup, out chan<- schema.CheckResult) {
	defer wg.Done()

	// Set up result
	result := schema.CheckResult{
		Timestamp:   time.Now(),
		ID:          d.ID,
		Name:        d.Name,
		Group:       d.Group,
		ScoreWeight: d.ScoreWeight,
		CheckType:   "vnc",
	}

	// Make channels for completing the check or not
	done := make(chan bool)
	failed := make(chan bool)

	go func() {
		// Configure the vnc client
		config := vnc.ClientConfig{
			Auth: []vnc.ClientAuth{
				&vnc.PasswordAuth{Password: d.Password},
			},
		}

		// Dial the vnc server
		// conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", d.Host, d.Port), 5*time.Second)
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", d.Host, d.Port))
		if err != nil {
			result.Message = fmt.Sprintf("Connection to VNC host %s failed : %s", d.Host, err)
			failed <- true
			return
		}
		defer conn.Close()

		vncClient, err := vnc.Client(conn, &config)
		if err != nil {
			result.Message = fmt.Sprintf("Login to server %s failed : %s", d.Host, err)
			failed <- true
			return
		}
		defer vncClient.Close()

		// If we made it here the check passes
		done <- true
		return
	}()

	// Watch channels and context for timeout
	for {
		select {
		case <-done:
			close(done)
			result.Passed = true
			out <- result
			return
		case <-failed:
			close(failed)
			out <- result
			return
		case <-ctx.Done():
			result.Message = fmt.Sprintf("Timeout via context : %s", ctx.Err())
			out <- result
			return
		}
	}
}

// Init the check using a known ID and name. The rest of the check fields will
// be filled in by parsing a JSON string representing the check definition.
func (d *Definition) Init(id string, name string, group string, scoreWeight float64, def []byte) error {

	// Unpack JSON definition
	err := json.Unmarshal(def, &d)
	if err != nil {
		return err
	}

	// Set generic values
	d.ID = id
	d.Name = name
	d.Group = group
	d.ScoreWeight = scoreWeight

	// Check for missing fields
	missingFields := make([]string, 0)
	if d.Host == "" {
		missingFields = append(missingFields, "Host")
	}

	if d.Port == "" {
		missingFields = append(missingFields, "Port")
	}

	if d.Password == "" {
		missingFields = append(missingFields, "Password")
	}

	// Error only the first missing field, if there are any
	if len(missingFields) > 0 {
		return schema.ValidationError{
			ID:    d.ID,
			Type:  "vnc",
			Field: missingFields[0],
		}
	}
	return nil
}
