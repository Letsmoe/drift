package neoworks

import (
	_ "embed"
)

// collection is the registry path of the collection Drift syncs.
const collection = "@letsmoe/drift"

// schemaVersion is the published version of the bundled descriptor, the
// version in schema/drift/openschema.yaml.
const schemaVersion = "0.1.0"

// descriptor is the published descriptor of the collection, byte for byte.
//
//go:embed schema/drift/generated/schema.descriptor.json
var descriptor []byte

// scopes are what Drift asks for at sign-in: read and write on its collection.
func scopes() []string {
	return []string{"openid", "profile", "email", collection + ":read", collection + ":write"}
}
