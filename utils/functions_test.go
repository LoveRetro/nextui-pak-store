package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/LoveRetro/nextui-pak-store/models"
)

func TestUnzipPakZipContentsIntoPakDirectory(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "Example.pak.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(archiveFile)
	entry, err := zw.Create("launch.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(entry, "#!/bin/sh\n"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(root, "Tools", "h700", "Example.pak")
	if err := Unzip(archivePath, destination, models.Pak{}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(destination, "launch.sh")); err != nil {
		t.Fatalf("launch.sh was not installed at the Pak root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, "Example.pak")); !os.IsNotExist(err) {
		t.Fatal("archive created a nested Example.pak directory")
	}
}
