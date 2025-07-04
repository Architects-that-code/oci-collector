package billing

import (
	"bufio"
	config "check-limits/config"
	"context"
	"fmt"
	"io"
	"log"
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

	for _, obj := range filesToDownload {
		wg.Add(1)
		go func(obj objectstorage.ObjectSummary) {
			defer wg.Done()

			fmt.Println("Downloading: ", *obj.Name)

			getReq := objectstorage.GetObjectRequest{
				NamespaceName: common.String(reportingNamespace),
				BucketName:    common.String(tenancyID),
				ObjectName:    common.String(*obj.Name),
			}

			objDetail, err := client.GetObject(context.Background(), getReq)
			if err != nil {
				log.Println("Error downloading file: ", *obj.Name, err)
				return
			}

			filename := *obj.Name
			filePath := filepath.Join(homedir, outputPath, profileName, filename)

			content, _ := io.ReadAll(objDetail.Content)

			if err := writeFile(filePath, content); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Downloaded file: %s\n", *obj.Name)

			mu.Lock()
			defer mu.Unlock()
			f, err := os.OpenFile(stateFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println("Error opening state file for append: ", err)
				return
			}
			defer f.Close()
			if _, err := f.WriteString(*obj.Name + "\n"); err != nil {
				log.Println("Error writing to state file: ", err)
			}
		}(obj)
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
