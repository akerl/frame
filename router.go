package main

import (
	"context"
	_ "embed"
	"math/rand"
	"net/http"
	"time"

	"github.com/akerl/go-lambda/apigw/events"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:embed assets/index.html
var indexFile string

var images = []string{}
var imageUpdateTime = time.Time{}

func indexHandler(req events.Request) (events.Response, error) {
	if !validAuthToken(req.Headers["X-API-Key"]) {
		return events.Fail("unauthorized")
	}

	cookie := &http.Cookie{
		Name:     "X-API-Key",
		Value:    req.Headers["X-API-Key"],
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   31536000,
		Domain:   req.Headers["Host"],
	}

	return events.Response{
		StatusCode: 200,
		Body:       indexFile,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
			"Set-Cookie":   cookie.String(),
		},
	}, nil
}

func validAuthCookie(req events.Request) (events.Response, error) {
	header := http.Header{}
	header.Add("Cookie", req.Headers["Cookie"])
	request := http.Request{Header: header}
	cookie, err := request.Cookie("X-API-Key")
	if err == http.ErrNoCookie {
		return events.Fail("no cookie found")
	} else if err != nil {
		return events.Fail("parsing cookie failed")
	}

	if !validAuthToken(cookie.Value) {
		return events.Fail("unauthorized")
	}
	return events.Succeed("")
}

func randomHandler(req events.Request) (events.Response, error) {
	if resp, err := validAuthCookie(req); err != nil {
		return resp, err
	}

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

func validAuthToken(token string) bool {
	for _, i := range c.AuthTokens {
		if i == token {
			return true
		}
	}
	return false
}

func getClient() (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func getImages(client *s3.Client) ([]string, error) {
	if imageUpdateTime.Add(time.Minute * 60).After(time.Now()) {
		return images, nil
	}

	paginator := s3.NewListObjectsV2Paginator(
		client,
		&s3.ListObjectsV2Input{Bucket: &c.ImageBucket},
	)
	images = []string{}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []string{}, err
		}
		for _, obj := range page.Contents {
			images = append(images, *obj.Key)
		}
	}
	imageUpdateTime = time.Now()
	return images, nil
}
