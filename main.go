package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "luty",
	Short: "Cross-platform file compression tool",
	Long:  `luty is a cross-platform command line tool that compresses folders into single .ltp files`,
}

var zipCmd = &cobra.Command{
	Use:   "z",
	Short: "Compress folder",
	Long:  `Compress a folder into a single .ltp file`,
	Args:  cobra.ExactArgs(1),
	Run:   zipRun,
}

var target string

func init() {
	rootCmd.AddCommand(zipCmd)
	zipCmd.Flags().StringVarP(&target, "target", "t", "", "Specify output file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func zipRun(_ *cobra.Command, args []string) {
	folderPath := args[0]

	// Check if folder exists
	info, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(os.Stderr, "Error: Folder does not exist")
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Error: Cannot access folder - %v\n", err)
		}
		os.Exit(1)
	}

	// Verify it's a directory
	if !info.IsDir() {
		_, _ = fmt.Fprintln(os.Stderr, "Error: Path is not a folder")
		os.Exit(1)
	}

	// Determine output filename
	var outputPath string
	if target != "" {
		// Check -t parameter extension
		if !strings.HasSuffix(strings.ToLower(target), ".ltp") {
			_, _ = fmt.Fprintln(os.Stderr, "Error: Target file must have .ltp extension")
			os.Exit(1)
		}
		outputPath = target
	} else {
		// Default output: foldername.ltp
		folderName := filepath.Base(folderPath)
		outputPath = folderName + ".ltp"
	}

	// Ensure output path is absolute
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: Cannot get current working directory - %v\n", err)
			os.Exit(1)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	// Create ZIP file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Cannot create output file - %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := zipFile.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: Failed to close file - %v\n", closeErr)
		}
	}()

	// Create ZIP writer
	w := zip.NewWriter(zipFile)
	defer func() {
		if err := w.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: Failed to close ZIP writer - %v\n", err)
		}
	}()

	// Get absolute path
	absFolderPath, err := filepath.Abs(folderPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Cannot get absolute path - %v\n", err)
		os.Exit(1)
	}

	// Recursively compress folder
	err = filepath.Walk(absFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Create relative path
		relPath, err := filepath.Rel(absFolderPath, path)
		if err != nil {
			return fmt.Errorf("error creating relative path for %s: %w", path, err)
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		// Create ZIP header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("error creating ZIP header for %s: %w", path, err)
		}

		// Use forward slashes for paths in ZIP (cross-platform compatibility)
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		// Add trailing slash for directories
		if info.IsDir() {
			header.Name += "/"
		}

		// Write header
		writer, err := w.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("error creating ZIP entry for %s: %w", path, err)
		}

		// If it's a file, write content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", path, err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to close file %s - %v\n", path, closeErr)
				}
			}()

			_, err = io.Copy(writer, file)
			if err != nil {
				return fmt.Errorf("error compressing file %s: %w", path, err)
			}
		}

		return nil
	})

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Compression failed - %v\n", err)
		// Try to clean up the partial ZIP file
		if removeErr := os.Remove(outputPath); removeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to remove partial file - %v\n", removeErr)
		}
		os.Exit(1)
	}

	// Show compression result
	fmt.Printf("Compression completed: %s -> %s\n", folderPath, outputPath)
}
