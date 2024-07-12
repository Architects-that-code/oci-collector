package oos

import (
	"context"
	"fmt"

	"strconv"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

func ObjectCount(namespace, bucketName string, objectStorageClient objectstorage.ObjectStorageClient) string {
	// Create a context for the API call
	ctx := context.Background()

	// Create the request to get the bucket metadata
	req := objectstorage.GetBucketRequest{
		NamespaceName:   &namespace,
		BucketName:      &bucketName,
		Fields:          []objectstorage.GetBucketFieldsEnum{objectstorage.GetBucketFieldsApproximatecount},
		RequestMetadata: common.RequestMetadata{},
	}

	// Call the API to get the bucket metadata
	res, err := objectStorageClient.GetBucket(ctx, req)
	if err != nil {
		fmt.Printf("Error getting bucket: %v\n", err)
	}
	//log.Printf("res: %v\n", res)

	// Get the object count from the bucket metadata
	objectCount := res.Bucket.ApproximateCount
	var size = *objectCount
	//fmt.Printf("bucket %v in region %v has approximately %s objects\n", bucketName, objectStorageClient.Endpoint(), strconv.FormatInt(int64(*objectCount), 10))
	return strconv.Itoa(int(size))
}
