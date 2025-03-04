package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var directories = []string{
	"./uploads/avatar",
	"./uploads/labbook",
}

// CleanUploads cleans up old or unnecessary uploads.
func CleanUploads() error {
	for _, dir := range directories {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		// Loop through and delete each file/folder inside the directory
		for _, entry := range entries {
			entryPath := filepath.Join(dir, entry.Name())

			err := os.RemoveAll(entryPath) // Delete file or subdirectory
			if err != nil {
				return fmt.Errorf("failed to remove %s: %w", entryPath, err)
			}
		}
	}

	return nil
}

func StartAutoCleanUploads(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := CleanUploads()
			if err != nil {
				fmt.Println("Error cleaing uploads:", err)
			}
		}
	}()
}

// CreateUploadsDir creates the necessary directories for file uploads.
func CreateUploadsDir() error {
	// Define the directories to be created

	// Loop through and create each directory
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return nil
}
