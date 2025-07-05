package billing

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ProcessBillingFiles scans the filesystem for gzipped billing files,
// decompresses them, combines their CSV content, and writes a single
// combined CSV file named by the earliest and latest dates found in the filenames.
func ProcessBillingFiles(outputPath, profileName string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	profilePath := filepath.Join(homedir, outputPath, profileName)

	// Ensure the output directory exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile directory does not exist: %s", profilePath)
	}

	// Setup error log file
	errorFilePath := filepath.Join(profilePath, "error_files.txt")
	errorFileHandle, err := os.OpenFile(errorFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening error log file: %w", err)
	}
	defer errorFileHandle.Close()
	errorWriter := bufio.NewWriter(errorFileHandle)
	defer errorWriter.Flush()
	var errorFileMutex sync.Mutex

	// Find all .gz files in the profile directory recursively
	var gzFiles []string
	err = filepath.WalkDir(profilePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".gz") {
			gzFiles = append(gzFiles, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if len(gzFiles) == 0 {
		fmt.Println("No .gz billing files found in", profilePath)
		return nil
	}

	var earliestDate, latestDate time.Time
	outputFileCreated := false
	tempOutputFilePath := filepath.Join(profilePath, "temp_combined_billing.csv")

	outputFile, err := os.Create(tempOutputFilePath)
	if err != nil {
		return fmt.Errorf("error creating temporary combined CSV file: %w", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	var wg sync.WaitGroup
	var dateMutex sync.Mutex
	var writerMutex sync.Mutex

	maxConcurrentProcessing := 25 // Limit to 25 concurrent processing tasks
	sem := make(chan struct{}, maxConcurrentProcessing)

	totalFiles := len(gzFiles)
	for i, gzFile := range gzFiles {
		wg.Add(1)
		sem <- struct{}{} // Acquire a token

		go func(index int, gzFile string) {
			defer wg.Done()
			defer func() { <-sem }() // Release the token when done

			relativePath, err := filepath.Rel(profilePath, gzFile)
			if err != nil {
				log.Printf("Error getting relative path for %s: %v\n", gzFile, err)
				relativePath = filepath.Base(gzFile) // Fallback to base name if relative path fails
			}
			fmt.Printf("Processing %d of %d: %s\n", index+1, totalFiles, relativePath)
			file, err := os.Open(gzFile)
			if err != nil {
				log.Printf("Error opening %s: %v\n", gzFile, err)
				errorFileMutex.Lock()
				fmt.Fprintln(errorWriter, relativePath)
				errorWriter.Flush()
				errorFileMutex.Unlock()
				return
			}
			defer file.Close()

			gzr, err := gzip.NewReader(file)
			if err != nil {
				log.Printf("Error creating gzip reader for %s: %v\n", gzFile, err)
				errorFileMutex.Lock()
				fmt.Fprintln(errorWriter, relativePath)
				errorWriter.Flush()
				errorFileMutex.Unlock()
				return
			}
			defer gzr.Close()

			reader := csv.NewReader(gzr)

			// Extract date from path for combined CSV name
			fileDate, err := extractDateFromPath(gzFile)
			if err != nil {
				log.Printf("Could not extract date from path %s: %v\n", gzFile, err)
				errorFileMutex.Lock()
				fmt.Fprintln(errorWriter, relativePath)
				errorWriter.Flush()
				errorFileMutex.Unlock()
				return
			}

			writerMutex.Lock()
			if !outputFileCreated {
				// Read and write header from the first file
				header, err := reader.Read()
				if err != nil {
					log.Printf("Error reading header from %s: %v\n", gzFile, err)
					errorFileMutex.Lock()
					fmt.Fprintln(errorWriter, relativePath)
					errorWriter.Flush()
					errorFileMutex.Unlock()
					writerMutex.Unlock()
					return
				}
				if err := writer.Write(header); err != nil {
					writerMutex.Unlock()
					return // Propagate error up
				}
				outputFileCreated = true
			} else {
				// Skip header for subsequent files
				_, err := reader.Read()
				if err != nil && err != io.EOF {
					log.Printf("Error skipping header in %s: %v\n", gzFile, err)
					errorFileMutex.Lock()
					fmt.Fprintln(errorWriter, relativePath)
					errorWriter.Flush()
					errorFileMutex.Unlock()
					writerMutex.Unlock()
					return
				}
			}
			writerMutex.Unlock()

			dateMutex.Lock()
			if earliestDate.IsZero() || fileDate.Before(earliestDate) {
				earliestDate = fileDate
			}
			if latestDate.IsZero() || fileDate.After(latestDate) {
				latestDate = fileDate
			}
			dateMutex.Unlock()

			// Write remaining records directly to the output file
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading record from %s: %v\n", gzFile, err)
					errorFileMutex.Lock()
					fmt.Fprintln(errorWriter, relativePath)
					errorWriter.Flush()
					errorFileMutex.Unlock()
					break // Change continue to break to move to the next file
				}
				writerMutex.Lock()
				if err := writer.Write(record); err != nil {
					writerMutex.Unlock()
					fmt.Errorf("error writing record to combined CSV: %w", err)

				}
				writerMutex.Unlock()
			}
		}(i, gzFile)
	}

	wg.Wait()

	// Check if any data was written to the combined file
	if !outputFileCreated {
		fmt.Println("No valid CSV data processed to combine.")
		os.Remove(tempOutputFilePath) // Clean up empty temp file
		return nil
	}

	// Rename the temporary file to the final dated name
	finalOutputFileName := fmt.Sprintf("%s-%s.csv", earliestDate.Format("2006-01-02"), latestDate.Format("2006-01-02"))
	finalOutputFilePath := filepath.Join(profilePath, finalOutputFileName)

	if err := os.Rename(tempOutputFilePath, finalOutputFilePath); err != nil {
		return fmt.Errorf("error renaming combined CSV file: %w", err)
	}

	fmt.Printf("Combined CSV written to: %s\n", finalOutputFilePath)
	return nil
}

// extractDateFromPath extracts a date (YYYY/MM/DD) from a given file path.
func extractDateFromPath(filePath string) (time.Time, error) {
	// Regex to find date patterns like YYYY/MM/DD in the path
	re := regexp.MustCompile(`(\d{4}/\d{2}/\d{2})`)
	matches := re.FindStringSubmatch(filePath)

	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("no date found in path: %s", filePath)
	}

	dateStr := matches[1]
	parsedDate, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse date from path: %s, error: %w", filePath, err)
	}

	return parsedDate, nil
}
