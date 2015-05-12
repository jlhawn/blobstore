package blobstore

import (
	"os"
)

// Ref increments the reference count for the blob with the given digest.
func (ls *localStore) Ref(digest string) (d Descriptor, err error) {
	ls.Lock()
	defer ls.Unlock()

	// Avoid the (type, nil) interface issue.
	d, blobErr := ls.ref(digest)
	if blobErr != nil {
		return nil, blobErr
	}

	return
}

// Deref decrements the reference count for the blob with the given digest. If
// no references to the blob remain, the blob will be removed from the store
// and blobs to which it links will also be dereferenced.
func (ls *localStore) Deref(digest string) error {
	ls.Lock()
	defer ls.Unlock()

	// Avoid the (type, nil) interface issue.
	blobErr := ls.deref(digest)
	if blobErr != nil {
		return blobErr
	}

	return nil
}

// ref is the unexported version of Ref which does not acquire the store lock
// before incrementing a blob reference count.
func (ls *localStore) ref(digest string) (Descriptor, *storeError) {
	info, err := ls.getBlobInfo(digest)
	if err != nil {
		return nil, err
	}

	info.RefCount++

	if err = ls.putBlobInfo(info); err != nil {
		return nil, err
	}

	return newDescriptor(info), nil
}

// deref is the unexported version of Deref which does not acquire the store
// lock before decrementing the blob reference count.
func (ls *localStore) deref(digest string) *storeError {
	info, err := ls.getBlobInfo(digest)
	if err != nil {
		return err
	}

	if info.RefCount > 0 { // Be careful to not overflow.
		info.RefCount--
	}

	if info.RefCount > 0 {
		return ls.putBlobInfo(info)
	}

	// Dereference blobs to which this one links.
	for _, linkTo := range info.LinksTo {
		if err = ls.deref(linkTo); !(err == nil || err.IsBlobNotExists()) {
			return err
		}
	}

	blobDirname := ls.blobDirname(digest)
	if e := os.RemoveAll(blobDirname); e != nil {
		return newError(errCodeCannotRemoveBlob, e.Error())
	}

	return nil
}
