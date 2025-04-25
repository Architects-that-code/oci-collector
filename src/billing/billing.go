package billing

import (
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

func GetBillingInfo() {
	fmt.Println("Getting billing info")
}

func Getfiles(provider common.ConfigurationProvider, tenancyID string, homeRegion string, config config.Config, outputPath string) {
	fmt.Println("Getting billing files")
	destinationPath := outputPath
	fmt.Println("destinationPath: ", destinationPath)
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	ctx := context.Background()

	// Create destination directory
	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		err = os.Mkdir(destinationPath, 0755)
		if err != nil {
			panic(err)
		}
	}

	var objSums []objectstorage.ObjectSummary
	// Get all report objects
	reportBucket := tenancyID
	listObjectRequest := objectstorage.ListObjectsRequest{
		NamespaceName: common.String(reportingNamespace),
		BucketName:    common.String(reportBucket),
		Prefix:        common.String(focusPrefixFile),
	}

	for {
		listObjectsResponse, err := client.ListObjects(ctx, listObjectRequest)
		helpers.FatalIfError(err)

		for _, objectSummary := range listObjectsResponse.ListObjects.Objects {
			objSums = append(objSums, objectSummary)
		}
		if listObjectsResponse.ListObjects.NextStartWith != nil {
			listObjectRequest.Start = listObjectsResponse.ListObjects.NextStartWith
		} else {
			break
		}
	}
	fmt.Println(len(objSums))
	// can we do the following in parallel?
	var wg sync.WaitGroup

	for _, obj := range objSums {
		wg.Add(1)
		go func(obj objectstorage.ObjectSummary) {
			defer wg.Done()

			fmt.Println(*obj.Name)

			getReq := objectstorage.GetObjectRequest{
				NamespaceName: common.String(reportingNamespace),
				BucketName:    common.String(reportBucket),
				ObjectName:    common.String(*obj.Name),
			}

			objDetail, err := client.GetObject(ctx, getReq)
			helpers.FatalIfError(err)

			// Extract filename
			filename := *obj.Name
			fmt.Println("filename: ", filename)

			// Download and save file
			homedir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			filePath := homedir + "/" + destinationPath + "/" + config.ProfileName + "/" + filename
			fmt.Println("orig file: ", filePath)
			fmt.Println("file size: ", *objDetail.ContentLength)
			//fmt.Println("contentType: ", *objDetail.ContentType)
			content, _ := io.ReadAll(objDetail.Content)
			//fmt.Println("content: ", content)

			if err := writeFile(filePath, content); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Downloaded file: %s\n", *obj.Name)
		}(obj)
	}

	wg.Wait()

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
