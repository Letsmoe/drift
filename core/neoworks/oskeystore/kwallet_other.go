//go:build !linux

package oskeystore

func platformSources() []keySource {
	return nil
}
