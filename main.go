package main

import (
	"archive/tar"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

// CreateDatedZstdTarball takes a source path and a target directory, creates a
// zstd-compressed tarball of the source, and saves it to the target directory.
// A CRC32 hash of the archive's content is always included in the filename
// (mm-dd-yyyy-crc32hash.tar.zstd) to ensure uniqueness for each revision.
// It returns true on success and false on any error.
func CreateDatedZstdTarball(sourcePath, targetDir string) bool {
	finalPath, err := createTarball(sourcePath, targetDir)
	if err != nil {
		log.Printf("Error creating tarball: %v", err)
		return false
	}
	log.Printf("Successfully created unique tarball: %s", finalPath)
	return true
}

// createTarball is the internal implementation that handles the logic and returns
// the final path of the created archive or a detailed error.
func createTarball(sourcePath, targetDir string) (string, error) {
	// 1. Validate source path
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source path '%s': %w", sourcePath, err)
	}
	if !sourceInfo.IsDir() {
		return "", fmt.Errorf("source path '%s' is not a directory", sourcePath)
	}

	// 2. Ensure the target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory '%s': %w", targetDir, err)
	}

	// 3. Create a temporary file to build the archive. This prevents partial files.
	tempFile, err := os.CreateTemp(targetDir, "backup-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file on error
	defer tempFile.Close()

	// 4. Set up the CRC32 hasher and the MultiWriter to write to both the
	// temp file and the hasher simultaneously.
	hasher := crc32.NewIEEE()
	multiWriter := io.MultiWriter(tempFile, hasher)

	// 5. Set up the chain of writers: file content -> tar -> zstd -> multiWriter
	zstdWriter, err := zstd.NewWriter(multiWriter,
		zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return "", fmt.Errorf("failed to create zstd writer: %w", err)
	}
	tarWriter := tar.NewWriter(zstdWriter)

	// 6. Walk the source directory and add files to the tarball.
	walkErr := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == sourcePath {
			return nil
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("could not create tar header for '%s': %w", path, err)
		}
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("could not calculate relative path for '%s': %w", path, err)
		}
		header.Name = filepath.ToSlash(relPath)
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("could not write tar header for '%s': %w", header.Name, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file '%s' for archiving: %w", path, err)
		}
		defer file.Close()
		if _, err := io.Copy(tarWriter, file); err != nil {
			return fmt.Errorf("could not copy file content from '%s' to tar archive: %w", path, err)
		}
		log.Printf("Added to archive: %s", header.Name)
		return nil
	})

	// 7. IMPORTANT: Close writers to flush all data before getting the hash.
	if err := tarWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := zstdWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close zstd writer: %w", err)
	}

	if walkErr != nil {
		return "", fmt.Errorf("error during directory walk: %w", walkErr)
	}

	// 8. Get the final hash and determine the unique, final filename.
	hash := hasher.Sum32()
	dateStr := time.Now().Format("01-02-2006")
	// Filename format is always: mm-dd-yyyy-crc32hash.tar.zstd
	finalFilename := fmt.Sprintf("%s-%x.tar.zstd", dateStr, hash)
	finalPath := filepath.Join(targetDir, finalFilename)

	// 9. Close the temp file and atomically rename it to its final destination.
	tempFile.Close()
	if err := os.Rename(tempFile.Name(), finalPath); err != nil {
		return "", fmt.Errorf("failed to rename temporary file to final path: %w", err)
	}

	return finalPath, nil
}

// main function to demonstrate usage.
func main() {
	const sourceDir = "/data"
	const targetDir = "/backups"

	log.Println("--- Starting Archive Process ---")
	success := CreateDatedZstdTarball(sourceDir, targetDir)
	if success {
		log.Println("--- Archive process completed successfully! ---")
	} else {
		log.Println("--- Archive process failed. ---")
	}

	// You could run it a second time with modified data to see a new file
	// with a different hash get created on the same day.

	// --- Clean up dummy files ---
	// _ = os.RemoveAll(sourceDir)
	// _ = os.RemoveAll(targetDir)
}
