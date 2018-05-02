package helper

import (
	"context"
	"path"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// ripped off shamelessly from vsphere tf provider internals

// PathIsEmpty checks a folder path to see if it's "empty" (ie: would resolve
// to the root inventory path for a given type in a datacenter - "" or "/").
func PathIsEmpty(path string) bool {
	return path == "" || path == "/"
}

// NormalizePath is a SchemaStateFunc that normalizes a folder path.
func NormalizePath(v interface{}) string {
	p := v.(string)
	if PathIsEmpty(p) {
		return ""
	}
	return strings.TrimPrefix(path.Clean(p), "/")
}

func FromAbsolutePath(client *govmomi.Client, path string) (*object.Folder, error) {
	finder := find.NewFinder(client.Client, false)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()
	folder, err := finder.Folder(ctx, path)
	if err != nil {
		return nil, err
	}
	return folder, nil
}
