// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
)

func Example_s3() {
	tr := instana.InitCollector(&instana.Options{
		Service: "s3-tracer",
	})

	s := tr.StartSpan("first-op")
	defer s.Finish()

	ctx := instana.ContextWithSpan(context.Background(), s)

	sdkConfig, err := config.LoadDefaultConfig(ctx)
	instaawsv2.Instrument(tr, &sdkConfig)

	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	count := 10
	fmt.Printf("Let's list up to %v buckets for your account.\n", count)

	var i int
	for i = 0; i < 10; i++ {

		result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err != nil {
			fmt.Printf("Couldn't list buckets for your account. Here's why: %v\n", err)
			return
		}
		if len(result.Buckets) == 0 {
			fmt.Println("You don't have any buckets!")
		} else {
			if count > len(result.Buckets) {
				count = len(result.Buckets)
			}
			for _, bucket := range result.Buckets[:count] {
				fmt.Printf("\t%v\n", *bucket.Name)
			}
		}

		time.Sleep(1 * time.Minute)
	}

}
