package neoworks

import (
	"context"
	"fmt"

	libneoworks "github.com/neoworks-dev/libneoworks-go"
)

// Sync pushes local changes to Neoworks and pulls remote ones; a failure
// libneoworks hit ends the run and comes back as the error. The first run
// of a session confirms with the registry that the bundled descriptors are
// the published ones; the store refuses the collections until then.
func (session *Session) Sync(ctx context.Context) (libneoworks.SyncReport, error) {
	if !session.SignedIn() {
		return libneoworks.SyncReport{}, ErrSignedOut
	}
	if err := session.registerDescriptors(ctx); err != nil {
		return libneoworks.SyncReport{}, err
	}
	report, err := session.client.Sync(ctx, transport{session: session})
	if err != nil {
		return report, fmt.Errorf("neoworks: sync: %w", err)
	}
	return report, syncFailure(report)
}

// syncFailure is the failure libneoworks reported, nil for a clean run. The
// report names the status even then ("ok"), so only its code tells.
func syncFailure(report libneoworks.SyncReport) error {
	if report.FailureCode == 0 {
		return nil
	}
	return fmt.Errorf("neoworks: sync: %w", report.FailureCode)
}

func (session *Session) registerDescriptors(ctx context.Context) error {
	session.mutex.Lock()
	registered := session.descriptorsRegistered
	session.mutex.Unlock()
	if registered {
		return nil
	}
	if err := session.registerDescriptor(ctx); err != nil {
		return err
	}
	session.mutex.Lock()
	session.descriptorsRegistered = true
	session.mutex.Unlock()
	return nil
}

func (session *Session) registerDescriptor(ctx context.Context) error {
	check, err := session.client.RegisterDescriptor(collection, descriptor, schemaVersion)
	if err != nil {
		return fmt.Errorf("neoworks: descriptor of %s: %w", collection, err)
	}
	if check == nil {
		return nil
	}
	response, err := transport{session: session}.Do(ctx, *check)
	if err != nil {
		return err
	}
	if err := session.client.CompleteDescriptorCheck(collection, response.Status, response.Body); err != nil {
		return fmt.Errorf("neoworks: the registry does not confirm %s: %w", collection, err)
	}
	return nil
}
