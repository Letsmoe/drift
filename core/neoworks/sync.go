package neoworks

import (
	"context"
	"fmt"

	libneoworks "github.com/neoworks-dev/libneoworks-go"
)

// Sync pushes local changes to Neoworks and pulls remote ones. The first run
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
	return report, nil
}

func (session *Session) registerDescriptors(ctx context.Context) error {
	session.mutex.Lock()
	registered := session.descriptorsRegistered
	session.mutex.Unlock()
	if registered {
		return nil
	}
	for _, name := range Collections {
		if err := session.registerDescriptor(ctx, name); err != nil {
			return err
		}
	}
	session.mutex.Lock()
	session.descriptorsRegistered = true
	session.mutex.Unlock()
	return nil
}

func (session *Session) registerDescriptor(ctx context.Context, name string) error {
	data, err := descriptor(name)
	if err != nil {
		return err
	}
	check, err := session.client.RegisterDescriptor(collectionPath(name), data)
	if err != nil {
		return fmt.Errorf("neoworks: descriptor of %s: %w", collectionPath(name), err)
	}
	if check == nil {
		return nil
	}
	response, err := transport{session: session}.Do(ctx, *check)
	if err != nil {
		return err
	}
	if err := session.client.CompleteDescriptorCheck(collectionPath(name), response.Status, response.Body); err != nil {
		return fmt.Errorf("neoworks: the registry does not confirm %s: %w", collectionPath(name), err)
	}
	return nil
}
