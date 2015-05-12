package blobstore

import (
	"testing"
)

func TestRefNotExists(t *testing.T) {
	withTestStore(t, testRefNotExists)
}

func testRefNotExists(t *testing.T, testStore *localStore) {
	_, err := testStore.Ref(randomDigest(t))
	if err == nil || !err.(Error).IsBlobNotExists() {
		t.Fatalf("expected %q error, got: %s", errDescriptions[errCodeBlobNotExists], err)
	}
}

func TestRef(t *testing.T) {
	withTestStore(t, testRef)
}

func testRef(t *testing.T, testStore *localStore) {
	numRefs := 5

	d1 := writeRandomBlob(t, testStore, 20480)

	if d1.RefCount() != 1 {
		t.Fatalf("expected reference count of 1, got %d", d1.RefCount())
	}

	for i := 1; i <= numRefs; i++ {
		d2, err := testStore.Ref(d1.Digest())
		if err != nil {
			t.Fatalf("unable to add reference %d: %s", 1, err)
		}

		ensureEqualDescriptors(t, d1, d2, false)

		if d2.RefCount() != uint64(i+1) {
			t.Fatalf("expected reference count of %d, got %d", i+1, d1.RefCount())
		}
	}
}

func TestDeref(t *testing.T) {
	withTestStore(t, testDeref)
}

func testDeref(t *testing.T, testStore *localStore) {
	numRefs := 5
	d1 := writeRandomBlob(t, testStore, 20480) // Start with refCount=1

	for i := 1; i < numRefs; i++ { // Add 4 more references.
		_, err := testStore.Ref(d1.Digest())
		if err != nil {
			t.Fatalf("unable to add reference %d: %s", i, err)
		}
	}

	for i := numRefs; i > 0; i-- { // Dereference 5 times.
		d2, err := testStore.Get(d1.Digest())
		if err != nil {
			t.Fatalf("unable to get blob %q: %s", d1.Digest(), err)
		}

		ensureEqualDescriptors(t, d1, d2, false)

		if d2.RefCount() != uint64(i) {
			t.Fatalf("expected refcount of %d, got %d", i, d2.RefCount())
		}

		if err := testStore.Deref(d1.Digest()); err != nil {
			t.Fatalf("unable to deref %q: %s", i, err)
		}
	}

	// The blob's references have all been removed
	// and the blob should no longer exist.
	_, err := testStore.Get(d1.Digest())
	if err == nil || !err.(Error).IsBlobNotExists() {
		t.Fatalf("expected %q error, got: %s", errDescriptions[errCodeBlobNotExists], err)
	}
}
