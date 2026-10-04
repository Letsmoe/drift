//go:build !linux

package browsercookies

import "errors"

func secretServicePassword(string) ([]byte, error) {
	return nil, errors.New("the Secret Service exists on Linux only")
}
