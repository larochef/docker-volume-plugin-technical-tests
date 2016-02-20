package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"encoding/json"
	"time"
	"github.com/docker/go-plugins-helpers/volume"
)

var definitions =  "definitions"
var data = "data"

type metadata struct {
	name string
        lo string
	created string
	size string
}

type myDriver struct {
	path       string
}

func newDriver(path string) myDriver {
	return myDriver{
		path: path,
	}
}

func (d myDriver) Create(r volume.Request) volume.Response {
	_, err := os.Open(filepath.Join(root, definitions, r.Name))
	if err != nil {
		return volume.Response{
			err: "Volume with name " + r.Name + " already exists",
		}
	}
	data := filepath.Join(root, data, r.Name)
	_, err2 := os.Open(data)
	if err2 != nil {
		return volume.Response{
			err: "Volume with name " + r.Name + " already exists",
		}
	}

	definition, e := os.Create(filepath.Join(root, definitions, r.Name))
	if e != nil {
		return volume.Response{
			err: "Unable to create file " + filepath.Join(root, definitions, r.Name) + ": " + e,
		}
	}

	_, e2 := os.Create(data)
	if e2 != nil {
		return volume.Response{
			err: "Unable to create file " + filepath.Join(root, definitions, r.Name) + ": " + e2,
		}
	}

	size := r.Options["size"]
	if size == "" {
		return volume.Response{
			err: "Size parameter is missing",
		}
	}
	now := time.Now
	lo := createData(data, r.Name, size)
	if lo == "" {
		return volume.Response{
			err: "Too many devices are mapped on host, remove some volumes before creating new ones.",
		}
	}
	createMetadata(definition, metadata {
		name: r.Name,
		lo: lo,
		created: now,
		size: size,
	})

	return nil
}

func createMetadata(file *os.File, metadata *metadata) error {
	bytes, err := json.Marshal(metadata)
	if(err == nil) {
		file.Write(bytes)
	}
	return err
}

func createData(file, name, size string) string {
	exec.Command ("dd", "if=/dev/zero", "of=" + file, "bs=1M", "count=" + size).Run()
	find := exec.Command ("losetup", "-f")
	find.Run()
	buf := new(bytes.Buffer)
	buf.ReadFrom(find.Stdin)
	lo := buf.String()

	if lo == "" {
		return ""
	}

	exec.Command ("losetup", lo, file).Run()
	exec.Command ("mkfs", "-t", "ext3", "-m", "1", "-v", lo).Run()
	return lo
}

func (d myDriver) Remove(r volume.Request) volume.Response {
	m, err := loadMetadata(r.Name)
	if err != nil {
		return volume.Response{err: err,}
	}
	exec.Command("losetup", "-d", m.lo)
	return nil
}

func (d myDriver) Path(r volume.Request) volume.Response {
	mntDir := filepath.Join(root, "mnt", r.Name)
	return volume.Response{
		Mountpoint: mntDir,
	}
}

func (d myDriver) Mount(r volume.Request) volume.Response {
	name := r.Name
	m, err := loadMetadata(name)
	if err != nil {
		return volume.Response{
			err: err,
		}
	}

	mntDir := filepath.Join(root, "mnt", name)
	os.Mkdir(mntDir, 755)
	exec.Command("mount", m.lo, mntDir)
	return volume.Response{
		Mountpoint: mntDir,
	}
}

func loadMetadata(name string) (*metadata, error) {
	file, err := os.Open(filepath.Join(root, definitions, name))
	if err != nil {
		return _, "Unable to find volume " + name + ": " + err
	}
	m := metadata{}
	err = json.NewDecoder(file).Decode(&m)
	if err != nil {
		return _, "Unable to read metadata for volume " + name + ": " + err
	}
	return m
}

func (d myDriver) Unmount(r volume.Request) volume.Response {
	name := r.Name
	mntDir := filepath.Join(root, "mnt", name)
	exec.Command("umount", mntDir).Run()
	return volume.Response{}
}

func (d myDriver) Get(r volume.Request) volume.Response {
	name := r.Name
	mntDir := filepath.Join(root, "mnt", name)

	return volume.Response{
		Volume: volume.Volume{
			Name: name,
			Mountpoint: mntDir,
		},
	}
}

func (d myDriver) List(r volume.Request) volume.Response {
	// files, _ := ioutil.ReadDir(filepath.Join(root, definitions))

	return nil
}
