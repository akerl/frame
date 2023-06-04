package main

import (
	"context"
	_ "embed"
	"math/rand"
	"time"

	"github.com/akerl/go-lambda/apigw/events"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:embed assets/index.html
var indexFile string

func indexHandler(_ events.Request) (events.Response, error) {
	return events.Response{
		StatusCode: 200,
		Body:       indexFile,
		Headers:    map[string]string{"Content-Type": "text/html; charset=utf-8"},
	}, nil
}

func randomHandler(_ events.Request) (events.Response, error) {
	client, err := getClient()
	if err != nil {
		return events.Fail("failed to load s3 client")
	}

	i, err := getImages(client)
	if err != nil {
		return events.Fail("failed to list images")
	}

	rand.Seed(time.Now().Unix())

	pc := s3.NewPresignClient(client)
	objReq, err := pc.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &c.ImageBucket,
		Key:    &i[rand.Intn(len(i))],
	})
	if err != nil {
		return events.Fail("Failed to load signed url")
	}
	return events.Redirect(objReq.URL, 303)
}

func getClient() (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func getImages(client *s3.Client) ([]string, error) {
	paginator := s3.NewListObjectsV2Paginator(
		client,
		&s3.ListObjectsV2Input{Bucket: &c.ImageBucket},
	)
	images := []string{}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []string{}, err
		}
		for _, obj := range page.Contents {
			images = append(images, *obj.Key)
		}
	}
	return images, nil
}
