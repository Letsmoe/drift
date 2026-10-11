package neoworks

import (
	"embed"
	"fmt"
)

//go:embed schema/*/generated/schema.descriptor.json
var descriptorFiles embed.FS

// Collections are the @drift collections Drift syncs, by name in the drift scope.
var Collections = []string{"tracks", "ratings", "saved", "playlists", "history", "blocks"}

// collectionPath is a collection's registry path, e.g. @drift/tracks.
func collectionPath(name string) string {
	return "@drift/" + name
}

// descriptor is the published descriptor of a collection, byte for byte.
func descriptor(name string) ([]byte, error) {
	data, err := descriptorFiles.ReadFile("schema/" + name + "/generated/schema.descriptor.json")
	if err != nil {
		return nil, fmt.Errorf("neoworks: descriptor of %s: %w", name, err)
	}
	return data, nil
}

// scopes are what Drift asks for at sign-in: read and write on its collections.
func scopes() []string {
	requested := []string{"openid", "profile", "email"}
	for _, name := range Collections {
		requested = append(requested, collectionPath(name)+":read", collectionPath(name)+":write")
	}
	return requested
}
