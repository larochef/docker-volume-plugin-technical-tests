package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "encoding/json"
    "github.com/docker/go-plugins-helpers/volume"
    "io/ioutil"
    "bufio"
    "bytes"
    "strings"
)

const definitions =  "definitions"
const data = "data"

type driverError struct {
    message string
}

func (error driverError) Error() string {
    return error.message
}


type metadata struct {
    Name string
    Lo string
    Created string
    Size string
}

type myDriver struct {
    path        string
    exposedPath string
}

func newDriver(path, exposedPath string) myDriver {
    return myDriver{
        path: path,
        exposedPath: exposedPath,
    }
}

func (d myDriver) Create(r volume.Request) volume.Response {

    println("Create volume with name", r.Name)

    _, err := os.Open(filepath.Join(d.path, definitions, r.Name))
    if err == nil {
        return volume.Response{
            Err: "Volume definition with name " + r.Name + " already exists",
        }
    }
    data := filepath.Join(d.path, data, r.Name)
    _, err2 := os.Open(data)
    if err2 == nil {
        return volume.Response{
            Err: "Volume data with name " + r.Name + " already exists",
        }
    }

    definition, e := os.Create(filepath.Join(d.path, definitions, r.Name))
    if e != nil {
        return volume.Response{
            Err: "Unable to create file " + filepath.Join(d.path, definitions, r.Name) + ": " + e.Error(),
        }
    }

    _, e2 := os.Create(data)
    if e2 != nil {
        return volume.Response{
            Err: "Unable to create file " + filepath.Join(d.path, definitions, r.Name) + ": " + e2.Error(),
        }
    }

    size := r.Options["size"]
    if size == "" {
        return volume.Response{
            Err: "Size parameter is missing",
        }
    }
    // now := time.Now
    lo := createData(data, r.Name, size)
    if lo == "" {
        return volume.Response{
            Err: "Too many devices are mapped on host, remove some volumes before creating new ones.",
        }
    }
    er := createMetadata(definition, &metadata {
        Name: r.Name,
        Lo: lo,
        Created: "",
        Size: size,
    })

    if er != nil {
        println("Error when writing metadata", er.Error())
    }

    return volume.Response{
    }
}

func createMetadata(file *os.File, metadata *metadata) error {
    bytes, err := json.Marshal(metadata)
    if(err == nil) {
        file.Write(bytes)
        file.Close()
    }
    return err
}

func createData(file, name, size string) string {
    exec.Command ("dd", "if=/dev/zero", "of=" + file, "bs=1M", "count=" + size).Run()
    find := exec.Command ("losetup", "-f")
    res, err := find.Output()
    if err != nil {
        println("Got error when executing losetup -f", err.Error())
        return ""
    }

    var lo string
    var e error

    lo, e = bufio.NewReader(bytes.NewReader(res)).ReadString('\n')
    lo = strings.Replace(lo, "\n", "", 5)

    if e != nil {
        println("An error occurred", e.Error())
        return ""
    }
    println("Mapping to local loop", lo)

    exec.Command ("losetup", lo, file).Run()
    exec.Command ("mkfs", "-t", "ext3", "-m", "1", "-v", lo).Run()
    return lo
}

func (d myDriver) Remove(r volume.Request) volume.Response {
    println("Removing volume", r.Name)

    m, err := loadMetadata(d.path, r.Name)
    if err != nil {
        return volume.Response{Err: err.Error(),}
    }
    // Discard errors that could occur if loop is already dissociated
    _ = exec.Command("unmount", m.Lo).Run()
    _ = exec.Command("losetup", "-d", m.Lo).Run()
    os.Remove(filepath.Join(d.path, data, m.Name))
    os.Remove(filepath.Join(d.path, definitions, m.Name))
    return volume.Response{}
}

func (d myDriver) Path(r volume.Request) volume.Response {
    _, err := loadMetadata(d.path, r.Name)

    if err != nil {
        return volume.Response{
            Err: "Unable to find volume " + r.Name,
        }
    }

    mntDir := filepath.Join(d.exposedPath, "mnt", r.Name)
    return volume.Response{
        Mountpoint: mntDir,
    }
}

func (d myDriver) Mount(r volume.Request) volume.Response {
    name := r.Name
    m, err := loadMetadata(d.path, name)
    if err != nil {
        return volume.Response{
            Err: err.Error(),
        }
    }

    mntDir := filepath.Join(d.path, "mnt", name)
    exposedPath := filepath.Join(d.exposedPath, "mnt", name)
    os.Mkdir(mntDir, 755)
    exec.Command("mount", m.Lo, mntDir).Run()
    return volume.Response{
        Mountpoint: exposedPath,
    }
}

func loadMetadata(path, name string) (*metadata, error) {
    file, err := os.Open(filepath.Join(path, definitions, name))
    if err != nil {
        return nil, driverError{ message: "Unable to find volume " + name + ": " + err.Error()}
    }
    m := metadata{}
    err = json.NewDecoder(file).Decode(&m)
    if err != nil {
        return nil, driverError{ message: "Unable to read metadata for volume " + name + ": " + err.Error()}
    }
    return &m, nil
}

func (d myDriver) Unmount(r volume.Request) volume.Response {
    name := r.Name
    mntDir := filepath.Join(d.path, "mnt", name)
    exec.Command("umount", mntDir).Run()
    return volume.Response{}
}

func (d myDriver) Get(r volume.Request) volume.Response {
    println("Inspecting volume", r.Name)

    name := r.Name

    _, err := loadMetadata(d.path, r.Name)

    if err == nil {
        return volume.Response{
            Volume: &volume.Volume{
                Name: name,
                Mountpoint: filepath.Join(d.exposedPath, "mnt", name),
            },
        }
    }

    return volume.Response{
        Err: "Volume " + name + " doesn't exist",
    }

}

func (d myDriver) List(r volume.Request) volume.Response {
    files, _ := ioutil.ReadDir(filepath.Join(d.path, definitions))
    result := make([]*volume.Volume, 0)
    for _, file := range files {
        _, name := filepath.Split(file.Name())

        volume := volume.Volume{
            Name: name,
            Mountpoint: filepath.Join(d.exposedPath, "mnt", name),
        }
        result = append(result, &volume, )
    }

    return volume.Response{
        Volumes: result,
    }
}
