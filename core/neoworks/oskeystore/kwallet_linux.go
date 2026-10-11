package oskeystore

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	kwalletAppID  = "Drift"
	kwalletFolder = "Drift"
)

// kwalletServices are the D-Bus names of KWallet 6 and 5 with their objects.
var kwalletServices = []struct {
	name string
	path dbus.ObjectPath
}{
	{"org.kde.kwalletd6", "/modules/kwalletd6"},
	{"org.kde.kwalletd5", "/modules/kwalletd5"},
}

// kwallet talks to an existing KWallet wallet over D-Bus. It never creates a
// wallet, which would open KWallet's setup dialog; without one it does not work.
type kwallet struct{}

func platformSources() []keySource {
	return []keySource{kwallet{}}
}

func (kwallet) name() string { return BackendKWallet }

func (wallet kwallet) load(account string) ([]byte, error) {
	var stored string
	err := wallet.withWallet(func(object dbus.BusObject, handle int32) error {
		var hasEntry bool
		err := object.Call("org.kde.KWallet.hasEntry", 0, handle, kwalletFolder, account, kwalletAppID).Store(&hasEntry)
		if err != nil {
			return fmt.Errorf("oskeystore: kwallet: %w", err)
		}
		if !hasEntry {
			return errKeyNotFound
		}
		err = object.Call("org.kde.KWallet.readPassword", 0, handle, kwalletFolder, account, kwalletAppID).Store(&stored)
		if err != nil {
			return fmt.Errorf("oskeystore: kwallet: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return decodeKey(stored)
}

func (wallet kwallet) store(account string, key []byte) error {
	return wallet.withWallet(func(object dbus.BusObject, handle int32) error {
		var result int32
		encoded := encodeKey(key)
		err := object.Call("org.kde.KWallet.writePassword", 0, handle, kwalletFolder, account, encoded,
			kwalletAppID).Store(&result)
		if err != nil {
			return fmt.Errorf("oskeystore: kwallet: %w", err)
		}
		if result != 0 {
			return fmt.Errorf("oskeystore: kwallet refused the key (%d)", result)
		}
		return nil
	})
}

// withWallet opens the network wallet when one exists and runs use on it.
func (kwallet) withWallet(use func(object dbus.BusObject, handle int32) error) error {
	connection, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("oskeystore: kwallet: %w", err)
	}
	for _, service := range kwalletServices {
		object := connection.Object(service.name, service.path)
		walletName, err := existingWallet(object)
		if err != nil {
			continue
		}
		var handle int32
		err = object.Call("org.kde.KWallet.open", 0, walletName, int64(0), kwalletAppID).Store(&handle)
		if err != nil || handle < 0 {
			return fmt.Errorf("oskeystore: kwallet did not open %s: %w", walletName, err)
		}
		defer object.Call("org.kde.KWallet.close", 0, handle, false, kwalletAppID)
		return use(object, handle)
	}
	return errors.New("oskeystore: no KWallet wallet")
}

// existingWallet names the wallet KWallet uses for passwords, when it exists.
func existingWallet(object dbus.BusObject) (string, error) {
	var wallets []string
	if err := object.Call("org.kde.KWallet.wallets", 0).Store(&wallets); err != nil {
		return "", fmt.Errorf("oskeystore: kwallet: %w", err)
	}
	var networkWallet string
	if err := object.Call("org.kde.KWallet.networkWallet", 0).Store(&networkWallet); err != nil {
		return "", fmt.Errorf("oskeystore: kwallet: %w", err)
	}
	for _, wallet := range wallets {
		if wallet == networkWallet {
			return wallet, nil
		}
	}
	return "", errors.New("oskeystore: KWallet has no wallet")
}
