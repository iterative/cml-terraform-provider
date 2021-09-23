// Code generated for package main by go-bindata DO NOT EDIT. (@generated)
// sources:
// images/provisioner/setup.sh
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _imagesProvisionerSetupSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x54\x5d\x4f\xdc\x3a\x10\x7d\xcf\xaf\x98\x9b\x8b\xd0\xbd\x0f\x4e\x00\xa1\xaa\x0d\xe2\x81\x8a\x0f\xa1\x22\xa8\x5a\x5e\x2a\x40\x2b\xc7\x9e\x24\xee\x26\xb6\xe5\x19\x2f\x8d\xd4\x1f\x5f\x65\x37\xbb\x9b\x55\x29\x2a\x0f\x79\xf1\x9c\x99\x39\x33\xe7\x64\xfe\xcd\x4b\x63\x73\x6a\x92\xcb\xeb\x9b\x8b\xd3\x7c\x21\x43\xde\xba\x3a\x57\x5d\x3b\x23\x96\x6a\x9e\xb5\xae\x4e\x4c\x05\x0f\xf0\x0f\x88\x0a\xd2\xbd\x01\x97\xc2\xd3\x09\x70\x83\x36\x01\x38\xbf\xf8\x78\x7d\x76\x3b\xbb\xfc\x72\x77\x7b\x7f\x71\x7b\x7e\x6a\x9d\x35\x96\x31\x48\xc5\x66\x81\x09\x00\xaa\xc6\x41\x7a\xf6\xf9\xbe\x28\xae\x90\x8b\xe2\x8c\x28\x76\x28\xbe\x21\xc1\x63\xca\x21\xe2\x63\x7a\x92\xc2\x4f\xa0\xa8\x1d\x30\x22\x08\x09\x39\xb2\xca\xa5\xe7\xe1\xcb\x94\xb3\x55\xa6\xf3\x0f\x07\x72\x99\xd9\x23\x25\x09\xac\xe0\xd2\x33\x04\xec\xdc\x02\x21\x5a\xc9\x8c\x56\xa3\x16\xd1\xd7\x41\x6a\xa4\x01\xd5\x13\x63\xa7\xb8\x05\x6d\x48\x96\x2d\x0e\x29\x42\x4b\xd3\xf6\x6b\x58\x46\x18\x16\x46\xe1\xb6\xa8\xd6\x62\x40\x05\xf4\x8e\x0c\xbb\xd0\x43\xb4\x66\x81\x81\x10\x44\xff\x0a\xca\x7b\x59\xd4\x86\x85\x72\x01\x73\xef\xe5\x14\xed\x19\xa2\xd7\x92\x11\xf6\xf7\x37\x4f\xa2\x46\x06\x63\x89\x65\xdb\x82\xe8\x81\x5c\xc5\xcf\x32\xa0\xf0\xc1\x79\x0c\x6c\x90\x84\x72\x5d\xe7\x2c\x94\xd1\xb4\x5a\x20\x11\x5a\x36\xb2\x85\xda\xf0\x86\xb0\x8a\xa1\x05\x51\xd1\xd7\x1b\x68\x98\x3d\x15\x79\x5e\x23\x67\xda\xa9\x39\x86\x4c\xb9\x0e\x84\x83\x1a\x59\x8c\x2f\xd4\x6c\x48\x50\xf3\x5b\x20\x01\x18\xeb\x46\xc2\xd0\x39\x0d\x42\x5e\xc1\x0a\x01\xb1\x8c\x96\xe3\xba\x31\x21\x57\x52\xb5\x20\x44\xe7\xb4\xa9\xfa\x65\x46\xb1\xc2\x14\xe1\x19\x96\x76\x0a\xd1\xe6\xeb\xfa\x4e\xcd\x07\xd6\x2f\x10\x1e\x84\x0e\xd8\xa2\x24\xa4\xac\x91\xd4\x18\xe5\x82\x1f\xc8\xe7\xb5\xaf\xd7\xf6\x18\x76\x36\xc7\x7e\x58\x3e\x88\xc9\x6e\xc5\xa0\xc6\x44\x89\x54\x63\x09\x0f\x32\xa8\xe6\x54\x76\xfa\xdd\xf1\xd3\x5f\xf4\x81\xbd\xff\x5a\x2a\x67\x63\x10\x84\xa2\xff\xa1\x93\xc6\xa6\x6f\xd3\x90\x31\x04\x59\xb9\xd0\x6d\x07\xa5\xed\x98\x1a\xcb\xcc\x3a\x8d\xe4\x62\x50\xb8\x1c\x8f\x90\xa3\x9f\x1d\x1e\x65\x3f\xd6\x53\x96\x92\x9a\xb7\x35\x1d\x4a\x7e\xdf\xfd\x2d\x26\xd1\x95\x20\x42\x87\xa5\x89\x47\x47\xad\xb1\xbb\x41\x90\x91\xdd\x98\x3a\x19\x00\xc4\x76\x04\xbb\x30\xda\xc8\xac\x36\xdc\xc4\x32\x33\x6e\x7c\x18\x3d\x34\xa8\x35\x28\xf4\x07\xc1\xde\x5e\x6f\xc5\xef\xf0\x7d\x76\x70\xbc\x1b\xc9\x5a\x43\x3c\xbd\x1b\x9b\xa3\xb1\x5a\x2e\x2d\x01\x99\x7e\x21\xeb\xf5\xe5\xee\x2c\x76\x9a\x7b\xb4\xf1\xfd\xe6\xaa\x04\x24\x96\x81\xc7\xdf\x23\x59\x9f\xbb\xbb\x4f\x53\x62\xe3\xd1\x4c\x2a\x93\xfc\x0a\x00\x00\xff\xff\x6d\xbb\xb9\x44\x6f\x05\x00\x00")

func imagesProvisionerSetupShBytes() ([]byte, error) {
	return bindataRead(
		_imagesProvisionerSetupSh,
		"images/provisioner/setup.sh",
	)
}

func imagesProvisionerSetupSh() (*asset, error) {
	bytes, err := imagesProvisionerSetupShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "images/provisioner/setup.sh", size: 1391, mode: os.FileMode(420), modTime: time.Unix(1631470006, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"images/provisioner/setup.sh": imagesProvisionerSetupSh,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"images": &bintree{nil, map[string]*bintree{
		"provisioner": &bintree{nil, map[string]*bintree{
			"setup.sh": &bintree{imagesProvisionerSetupSh, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
