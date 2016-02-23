package main

import (
//    "flag"
    "fmt"
    "os"
    "path/filepath"
    "github.com/docker/go-plugins-helpers/volume"
)

// const pluginID = "_simpleFS"

// var (
    // defaultDir   = "/tmp/volumes" // filepath.Join(volume.DefaultDockerRootDirectory, pluginID)
    // root *string = flag.String("root", defaultDir, "Storage directory")
// )

var root string

func main() {
    // flag.Parse()
    // root = flag.String("root", defaultDir, "Storage directory")
    root = "/tmp/volumes"
    // Base directory to store all
    os.MkdirAll(root, 755)
    // Metadata on volumes
    os.Mkdir(filepath.Join(root, "definitions"), 755)
    // Data of volumes
    os.Mkdir(filepath.Join(root, "data"), 755)
    os.Mkdir(filepath.Join(root, "mnt"), 755)
    fmt.Println(volume.NewHandler(newDriver(root)).ServeUnix("root", "simpleFS"))
}