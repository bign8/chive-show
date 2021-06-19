package datastore

import (
	"testing"

	"cloud.google.com/go/datastore"
)

func TestEntityKeysPropertyLoadSaver(t *testing.T) {
	keys := entityKeys{
		{IntID: 123}, // ensure dedupe
	}
	keys.addKeys([]*datastore.Key{
		datastore.IDKey(``, 123, nil),
		datastore.NameKey(``, `456`, nil),
	})
	props, err := keys.Save()
	if err != nil {
		t.Fatalf("Unable to serialize keys: %v", err)
	}
	empty := entityKeys{}
	if err = empty.Load(props); err != nil {
		t.Fatalf("Unable to Deserialize keys: %v", err)
	}
	if len(empty) != 2 {
		t.Fatalf("Needed 2 keys, got %d", len(empty))
	}
	flat := empty.toKeys(``)
	if flat[0].ID != 123 {
		t.Errorf("Wanted 123, got %d", flat[0].ID)
	}
	if flat[1].Name != "456" {
		t.Errorf("Wanted 456, Got %q", flat[1].Name)
	}
}
