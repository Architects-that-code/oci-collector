package oos

import (
	"context"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

func GetObjectStorageInfo(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartment []identity.Compartment, capacityFetch bool, capacityShapeType string) {
	fmt.Println("Getting object storage info")
	ctx := context.Background()
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	//fmt.Printf("namespace: %v\n", namespace(ctx, client))

	BucketInfo(provider, regions, namespace(ctx, client), tenancyID, compartment)

}

func ObjectStorageSize(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartment []identity.Compartment, capacityFetch bool, homeRegion string) {
	fmt.Println("Getting object storage info ObjectStorageSize")
	ctx := context.Background()
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	ns := namespace(ctx, client)
	fmt.Printf("namespace: %v\n", ns)

	BucketInfo(provider, regions, ns, tenancyID, compartment)

}

func namespace(ctx context.Context, c objectstorage.ObjectStorageClient) string {
	request := objectstorage.GetNamespaceRequest{}
	r, err := c.GetNamespace(ctx, request)
	helpers.FatalIfError(err)
	fmt.Println("getting namespace")
	return *r.Value
}

func BucketInfo(provider common.ConfigurationProvider, regions []identity.RegionSubscription, namespace string, tenancyId string, compartments []identity.Compartment) {

	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	//for each region for each compartent get list of buckets

	allBuckets := []objectstorage.BucketSummary{}
	var wg sync.WaitGroup
	wg.Add(len(regions))
	var regionalSlices = make(chan []objectstorage.BucketSummary, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			for _, compartment := range compartments {
				buckets := Buckets(client, namespace, *region.RegionName, compartment, tenancyId)
				allBuckets = append(allBuckets, buckets.Items...)
				//fmt.Printf("region: \t%v  \tcomp:%v \t\t\n", *region.RegionName, *compartment.Name)
			}
			regionalSlices <- allBuckets
		}(region)
	}
	wg.Wait()
	close(regionalSlices)

	fmt.Printf("all buckets: %v\n", len(allBuckets))
	/*
		for _, bucket := range allBuckets {

			fmt.Printf("bucket: %v freeformTags: %v  defined tags: %v \n", *bucket.Name, bucket.FreeformTags, bucket.DefinedTags)
			//fmt.Printf("bucket: %v\n", bucket)
		}
	*/

}

func Buckets(client objectstorage.ObjectStorageClient, namespace string, region string, compartment identity.Compartment, tenancyId string) objectstorage.ListBucketsResponse {
	client.SetRegion(region)

	fields := []objectstorage.ListBucketsFieldsEnum{
		objectstorage.ListBucketsFieldsTags,
	}

	request := objectstorage.ListBucketsRequest{
		NamespaceName: common.String(namespace),
		CompartmentId: common.String(*compartment.Id),
		Fields:        fields,
	}
	r, err := client.ListBuckets(context.Background(), request)

	helpers.FatalIfError(err)
	for _, bucket := range r.Items {
		fmt.Printf("bucket: %v \tapprox size: %v\n", *bucket.Name, GetObjectCount(namespace, *bucket.Name, client))
	}
	return r
}

func GetObjectCount(namespace, bucketName string, objectStorageClient objectstorage.ObjectStorageClient) int64 {
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
	return /*strconv.Itoa(int(size))*/ size
}
