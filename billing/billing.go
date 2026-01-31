package billing

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	config "oci-collector/config"
	"os"
	"path/filepath"
	"sync"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

const (
	reportingNamespace = "bling"
	//destinationPath    = "/Users/jscanlon/reports"
	//defaultPrefixFile = "" // Download all files
	// Uncomment for specific reports
	// costPrefixFile    = "reports/cost-csv"
	// usagePrefixFile   = "reports/usage-csv"
	focusPrefixFile = "FOCUS"
)

func Getfiles(provider common.ConfigurationProvider, tenancyID string, homeRegion string, config config.Config, outputPath string, download bool) {
	fmt.Println("Getting billing files")

	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	// Get all available billing files
	allFiles, err := getAllBillingFiles(client, tenancyID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Total files available: ", len(allFiles))

	// Write the full list of files to a file
	if err := writeFullFileList(allFiles, outputPath, config.ProfileName); err != nil {
		log.Fatal(err)
	}

	// Determine which files to download
	filesToDownload, err := getFilesToDownload(allFiles, outputPath, config.ProfileName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("New files to download: ", len(filesToDownload))

	if download {
		// Download the new files
		if len(filesToDownload) > 0 {
			fmt.Println("Starting download of new files...")
			downloadBillingFiles(client, tenancyID, filesToDownload, outputPath, config.ProfileName)
		}
	} else {
		fmt.Println("Download flag is set to false, skipping download.")
	}
}

func writeFullFileList(allFiles []objectstorage.ObjectSummary, outputPath, profileName string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fullListPath := filepath.Join(homedir, outputPath, profileName, "full_billing_file_list.txt")

	// Create the directory if it doesn't exist
	dirPath := filepath.Dir(fullListPath)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return err
	}

	file, err := os.Create(fullListPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, obj := range allFiles {
		fmt.Fprintln(w, *obj.Name)
	}
	return w.Flush()
}

func getAllBillingFiles(client objectstorage.ObjectStorageClient, tenancyID string) ([]objectstorage.ObjectSummary, error) {
	var objSums []objectstorage.ObjectSummary
	reportBucket := tenancyID
	listObjectRequest := objectstorage.ListObjectsRequest{
		NamespaceName: common.String(reportingNamespace),
		BucketName:    common.String(reportBucket),
		Prefix:        common.String(focusPrefixFile),
	}

	for {
		listObjectsResponse, err := client.ListObjects(context.Background(), listObjectRequest)
		if err != nil {
			return nil, err
		}

		for _, objectSummary := range listObjectsResponse.ListObjects.Objects {
			objSums = append(objSums, objectSummary)
		}
		if listObjectsResponse.ListObjects.NextStartWith != nil {
			listObjectRequest.Start = listObjectsResponse.ListObjects.NextStartWith
		} else {
			break
		}
	}
	return objSums, nil
}

func getFilesToDownload(allFiles []objectstorage.ObjectSummary, outputPath, profileName string) ([]objectstorage.ObjectSummary, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	stateFilePath := filepath.Join(homedir, outputPath, profileName, "downloaded_files.txt")

	// Create the directory if it doesn't exist
	dirPath := filepath.Dir(stateFilePath)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return nil, err
	}

	downloadedFiles := make(map[string]bool)
	if _, err := os.Stat(stateFilePath); !os.IsNotExist(err) {
		file, err := os.Open(stateFilePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			downloadedFiles[scanner.Text()] = true
		}
	}

	var filesToDownload []objectstorage.ObjectSummary
	for _, obj := range allFiles {
		if !downloadedFiles[*obj.Name] {
			filesToDownload = append(filesToDownload, obj)
		}
	}
	return filesToDownload, nil
}

func downloadBillingFiles(client objectstorage.ObjectStorageClient, tenancyID string, filesToDownload []objectstorage.ObjectSummary, outputPath, profileName string) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	homedir, _ := os.UserHomeDir()
	stateFilePath := filepath.Join(homedir, outputPath, profileName, "downloaded_files.txt")
	errorFilePath := filepath.Join(homedir, outputPath, profileName, "error_files.txt")

	// Create the directory for the error file if it doesn't exist
	errorFileDirPath := filepath.Dir(errorFilePath)
	if err := os.MkdirAll(errorFileDirPath, 0777); err != nil {
		log.Printf("Error creating directory for error file: %v\n", err)
		return
	}

	// Open error_files.txt for appending
	errorFile, err := os.OpenFile(errorFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening error_files.txt for writing: %v\n", err)
		return
	}
	defer errorFile.Close()

	totalFiles := len(filesToDownload)

	// Worker pool setup
	maxConcurrentDownloads := 25 // Limit to 5 concurrent downloads
	sem := make(chan struct{}, maxConcurrentDownloads)

	for i, obj := range filesToDownload {
		wg.Add(1)
		sem <- struct{}{} // Acquire a token

		go func(index int, obj objectstorage.ObjectSummary) {
			defer wg.Done()
			defer func() { <-sem }() // Release the token when done

			fmt.Printf("Downloading %d of %d: %s\n", index+1, totalFiles, *obj.Name)

			getReq := objectstorage.GetObjectRequest{
				NamespaceName: common.String(reportingNamespace),
				BucketName:    common.String(tenancyID),
				ObjectName:    common.String(*obj.Name),
			}

			objDetail, err := client.GetObject(context.Background(), getReq)
			if err != nil {
				log.Printf("Error downloading file %s: %v\n", *obj.Name, err)
				mu.Lock()
				fmt.Fprintln(errorFile, *obj.Name)
				mu.Unlock()
				return
			}

			filename := *obj.Name
			filePath := filepath.Join(homedir, outputPath, profileName, filename)

			content, err := io.ReadAll(objDetail.Content)
			if err != nil {
				log.Printf("Error reading downloaded content for %s: %v\n", *obj.Name, err)
				mu.Lock()
				fmt.Fprintln(errorFile, *obj.Name)
				mu.Unlock()
				return
			}

			// Validate MD5 hash
			if obj.Md5 != nil {
				h := md5.New()
				h.Write(content)
				calculatedMd5 := base64.StdEncoding.EncodeToString(h.Sum(nil))
				if calculatedMd5 != *obj.Md5 {
					log.Printf("MD5 mismatch for %s. Expected %s, got %s\n", *obj.Name, *obj.Md5, calculatedMd5)
					mu.Lock()
					fmt.Fprintln(errorFile, *obj.Name)
					mu.Unlock()
					return
				}
			}

			// Validate Gzip integrity (attempt to decompress)
			gzr, err := gzip.NewReader(bytes.NewReader(content))
			if err != nil {
				log.Printf("Gzip integrity check failed for %s: %v\n", *obj.Name, err)
				mu.Lock()
				fmt.Fprintln(errorFile, *obj.Name)
				mu.Unlock()
				return
			}
			gzr.Close()

			if err := writeFile(filePath, content); err != nil {
				log.Fatal(err)
			}

			mu.Lock()
			// Open downloaded_files.txt for appending
			f, err := os.OpenFile(stateFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println("Error opening state file for append: ", err)
				mu.Unlock()
				return
			}
			defer f.Close()
			if _, err := f.WriteString(*obj.Name + "\n"); err != nil {
				log.Println("Error writing to state file: ", err)
			}
			mu.Unlock()
		}(i, obj)
	}

	wg.Wait()
	fmt.Println("Billing file download process complete.")
}

func writeFile(path string, content []byte) error {
	// Create the directory if it doesn't exist

	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return err
	}

	// Write the file with default permissions
	return os.WriteFile(path, content, 0777)
}

func RedownloadErrorFiles(provider common.ConfigurationProvider, tenancyID string, homeRegion string, config config.Config, outputPath string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	errorFilePath := filepath.Join(homedir, outputPath, config.ProfileName, "error_files.txt")

	// Check if error_files.txt exists
	if _, err := os.Stat(errorFilePath); os.IsNotExist(err) {
		fmt.Println("No error_files.txt found. Nothing to re-download.")
		return nil
	}

	// Read files to re-download first
	var filesToRedownload []objectstorage.ObjectSummary
	file, err := os.Open(errorFilePath)
	if err != nil {
		return fmt.Errorf("error opening error_files.txt: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		objectName := scanner.Text()
		filesToRedownload = append(filesToRedownload, objectstorage.ObjectSummary{
			Name: common.String(objectName),
		})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading error_files.txt: %w", err)
	}

	// Now truncate the error file for the new re-download attempt
	truncateFile, err := os.OpenFile(errorFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening error_files.txt for truncation: %w", err)
	}
	truncateFile.Close() // Close immediately after truncation

	if len(filesToRedownload) == 0 {
		fmt.Println("error_files.txt is empty. Nothing to re-download.")
		return nil
	}

	fmt.Printf("Attempting to re-download %d error files.\n", len(filesToRedownload))

	// Create Object Storage client here
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	// Call the existing download function
	downloadBillingFiles(client, tenancyID, filesToRedownload, outputPath, config.ProfileName)

	fmt.Println("Re-download of error files complete. Please re-run processing to include these files.")
	return nil
}
