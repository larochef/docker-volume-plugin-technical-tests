package main

import (
    "os"
    "flag"
    "path/filepath"
    "github.com/docker/go-plugins-helpers/volume"
)

const defaultPath = "/tmp/volumes"
const pluginName = "simple-fs"

func main() {
    defer os.Remove("/run/docker/" + pluginName)

    root :=  flag.String("root", defaultPath, "Simple FS storage location")
    path :=  flag.String("path", filepath.Join(defaultPath, "mnt"), "Simple FS storage location")
    flag.Parse()
    // Base directory to store all
    os.MkdirAll(*root, 0755)
    // Metadata on volumes
    os.Mkdir(filepath.Join(*root, "definitions"), 0755)
    // Data of volumes
    os.Mkdir(filepath.Join(*root, "data"), 0755)
    os.Mkdir(filepath.Join(*root, "mnt"), 0755)
    println("Creating server from root dir", *root)
    var driver volume.Driver = newDriver(*root, *path)
    volume.NewHandler(driver).ServeUnix("root", pluginName)
}