//go:build linux

package browsercookies

import (
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	secretsBus          = "org.freedesktop.secrets"
	secretsPath         = dbus.ObjectPath("/org/freedesktop/secrets")
	defaultCollection   = dbus.ObjectPath("/org/freedesktop/secrets/aliases/default")
	noObject            = dbus.ObjectPath("/")
	secretServiceIface  = "org.freedesktop.Secret.Service"
	secretCollectionAPI = "org.freedesktop.Secret.Collection"
	secretItemAPI       = "org.freedesktop.Secret.Item"
	secretPromptAPI     = "org.freedesktop.Secret.Prompt"
	unlockPromptTimeout = 60 * time.Second
)

// secret is the (session, parameters, value, content_type) struct of GetSecret.
type secret struct {
	Session     dbus.ObjectPath
	Parameters  []byte
	Value       []byte
	ContentType string
}

// secretServicePassword reads the secret of the default collection's item
// with the given label, unlocking the collection when the user allows it.
func secretServicePassword(label string) ([]byte, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("connect to the session bus: %w", err)
	}
	service := conn.Object(secretsBus, secretsPath)
	var output dbus.Variant
	var session dbus.ObjectPath
	err = service.Call(secretServiceIface+".OpenSession", 0, "plain", dbus.MakeVariant("")).Store(&output, &session)
	if err != nil {
		return nil, fmt.Errorf("open a Secret Service session: %w", err)
	}
	defer conn.Object(secretsBus, session).Call("org.freedesktop.Secret.Session.Close", 0)

	err = unlockCollection(conn, service)
	if err != nil {
		return nil, err
	}
	itemPath, err := findItemByLabel(conn, label)
	if err != nil {
		return nil, err
	}
	var found secret
	err = conn.Object(secretsBus, itemPath).Call(secretItemAPI+".GetSecret", 0, session).Store(&found)
	if err != nil {
		return nil, fmt.Errorf("read the %q secret: %w", label, err)
	}
	return found.Value, nil
}

func unlockCollection(conn *dbus.Conn, service dbus.BusObject) error {
	collection := conn.Object(secretsBus, defaultCollection)
	locked, err := collection.GetProperty(secretCollectionAPI + ".Locked")
	if err != nil {
		return fmt.Errorf("read the default keyring: %w", err)
	}
	if isLocked, _ := locked.Value().(bool); !isLocked {
		return nil
	}
	var unlocked []dbus.ObjectPath
	var prompt dbus.ObjectPath
	err = service.Call(secretServiceIface+".Unlock", 0, []dbus.ObjectPath{defaultCollection}).Store(&unlocked, &prompt)
	if err != nil {
		return fmt.Errorf("unlock the default keyring: %w", err)
	}
	if prompt == noObject {
		return nil
	}
	return waitForPrompt(conn, prompt)
}

// waitForPrompt shows the keyring's unlock dialog and waits for its answer.
func waitForPrompt(conn *dbus.Conn, prompt dbus.ObjectPath) error {
	signals := make(chan *dbus.Signal, 4)
	conn.Signal(signals)
	defer conn.RemoveSignal(signals)
	err := conn.AddMatchSignal(dbus.WithMatchObjectPath(prompt), dbus.WithMatchInterface(secretPromptAPI), dbus.WithMatchMember("Completed"))
	if err != nil {
		return fmt.Errorf("watch the unlock prompt: %w", err)
	}
	err = conn.Object(secretsBus, prompt).Call(secretPromptAPI+".Prompt", 0, "").Err
	if err != nil {
		return fmt.Errorf("show the unlock prompt: %w", err)
	}
	timeout := time.After(unlockPromptTimeout)
	for {
		select {
		case signal := <-signals:
			if signal.Path == prompt && len(signal.Body) > 0 {
				dismissed, _ := signal.Body[0].(bool)
				return promptResult(dismissed)
			}
		case <-timeout:
			return errors.New("the keyring unlock prompt timed out")
		}
	}
}

func promptResult(dismissed bool) error {
	if dismissed {
		return errors.New("the keyring unlock prompt was dismissed")
	}
	return nil
}

func findItemByLabel(conn *dbus.Conn, label string) (dbus.ObjectPath, error) {
	collection := conn.Object(secretsBus, defaultCollection)
	items, err := collection.GetProperty(secretCollectionAPI + ".Items")
	if err != nil {
		return "", fmt.Errorf("list the default keyring: %w", err)
	}
	paths, _ := items.Value().([]dbus.ObjectPath)
	for _, path := range paths {
		itemLabel, err := conn.Object(secretsBus, path).GetProperty(secretItemAPI + ".Label")
		if err != nil {
			continue
		}
		if text, _ := itemLabel.Value().(string); text == label {
			return path, nil
		}
	}
	return "", fmt.Errorf("no %q entry in the default keyring", label)
}
