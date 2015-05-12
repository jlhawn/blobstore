package blobstore

// Link sets up a link between the blobs referenced by digests *to* and
// *from*. Assumes an implicit reference from *from* to *to*.
func (ls *localStore) Link(to, from string) error {
	ls.Lock()
	defer ls.Unlock()

	// Avoid the (type, nil) interface issue.
	if err := ls.link(to, from); err != nil {
		return err
	}

	return nil
}

// Unlink removes a link between the blobs referenced by digests *to* and
// *from*. Also dereferences the *to* blob.
func (ls *localStore) Unlink(to, from string) error {
	ls.Lock()
	defer ls.Unlock()

	// Avoid the (type, nil) interface issue.
	if err := ls.unlink(to, from); err != nil {
		return err
	}

	return nil
}

// link is the unexported version of Link which does not acquire the store lock
// before adding a link count.
func (ls *localStore) link(to, from string) *storeError {
	infoFrom, err := ls.getBlobInfo(from)
	if err != nil {
		return err
	}

	infoTo, err := ls.getBlobInfo(to)
	if err != nil {
		return err // Link target not found?
	}

	var linkToFound bool
	for _, linkTo := range infoFrom.LinksTo {
		if linkToFound = linkTo == to; linkToFound {
			break
		}
	}

	if linkToFound {
		// Need to discount a reference.
		if infoTo.RefCount > 0 { // Be careful to not overflow.
			infoTo.RefCount--
			return ls.putBlobInfo(infoTo)
		}
	} else {
		// Add the link to the blob info.
		infoFrom.LinksTo = append(infoFrom.LinksTo, to)
		return ls.putBlobInfo(infoFrom)
	}

	return nil
}

// unlink is the unexported version of Unlink which does not acquire the store
// lock before  removing a link.
func (ls *localStore) unlink(to, from string) (err *storeError) {
	infoFrom, err := ls.getBlobInfo(from)
	if err != nil {
		return err
	}

	if _, err = ls.getBlobInfo(to); err != nil {
		return err // Link target not found?
	}

	toIdx := -1
	for idx, linkTo := range infoFrom.LinksTo {
		if linkTo == to {
			toIdx = idx
			break
		}
	}

	if toIdx >= 0 {
		// The link was found.
		defer func() {
			if err == nil {
				// Need to deref the link target.
				err = ls.deref(to)
			}
		}()

		// Remove the link element.
		linksLen := len(infoFrom.LinksTo)
		if linksLen == 1 {
			// We're removing the only link.
			infoFrom.LinksTo = nil
		} else {
			// Swap with the last element of the list, then shrink the list.
			lastIdx := linksLen - 1
			infoFrom.LinksTo[lastIdx], infoFrom.LinksTo[toIdx] = infoFrom.LinksTo[toIdx], infoFrom.LinksTo[lastIdx]
			infoFrom.LinksTo = infoFrom.LinksTo[:lastIdx]
		}

		return ls.putBlobInfo(infoFrom)
	}

	return nil
}
