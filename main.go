package main

import (
    "os"
    "path/filepath"
    "github.com/docker/go-plugins-helpers/volume"
)

func main() {
    root := "/tmp/volumes"
    // Base directory to store all
    os.MkdirAll(root, 755)
    // Metadata on volumes
    os.Mkdir(filepath.Join(root, "definitions"), 755)
    // Data of volumes
    os.Mkdir(filepath.Join(root, "data"), 755)
    os.Mkdir(filepath.Join(root, "mnt"), 755)
    println("Creating server from root dir", root)
    var driver volume.Driver = newDriver(root)
    volume.NewHandler(driver).ServeUnix("root", "simpleFS")
    println("Server created, waiting for requests")
}