package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"math"
	"os"
	"path"
	"strings"
)

var (
	parallelism int    // # goroutines to divide the downloads among
	downloadDir string // where downloaded .spec.yamls will be saved
	bucket      string // s3 bucket where build cache lives
	prefix      string // s3 prefix under which build cache lives
	region      string // s3 region where bucket is located
)

func main() {
	initGlobalSettings()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		errf("error: %v", err)
	}

	svc := s3.New(sess)

	in := &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: aws.String("/"),
	}

	var tok string
	var keys []string
	for {
		if tok != "" {
			in.ContinuationToken = &tok
		}
		r, err := svc.ListObjectsV2(in)
		if err != nil {
			errf("Unable to list objects, %v", err)
		}

		for _, o := range r.Contents {
			k := aws.StringValue(o.Key)
			if strings.HasSuffix(k, ".spec.json") || strings.HasSuffix(k, ".spec.yaml") {
				keys = append(keys, k)
			}
		}

		if aws.BoolValue(r.IsTruncated) {
			tok = aws.StringValue(r.NextContinuationToken)
			continue
		}
		break
	}

	nkeys := len(keys)
	fmt.Println("# specs =", nkeys)

	downloader := s3manager.NewDownloader(sess)

	done := make(chan bool)
	dt := int(math.Ceil(float64(nkeys) / float64(parallelism)))
	for i := 0; i < parallelism; i += 1 {
		i1 := i * dt
		i2 := lower(i1+dt, nkeys)
		go downloadS3Objects(downloader, keys[i1:i2], done)
	}

	for i := 0; i < parallelism; i += 1 {
		<-done
	}
}

func initGlobalSettings() {
	flag.IntVar(&parallelism, "n", 1, "degree of parallelism")
	flag.StringVar(&bucket, "b", "", "S3 bucket")
	flag.StringVar(&prefix, "p", "build_cache/", "prefix to build cache")
	flag.StringVar(&downloadDir, "d", "", "directory where downloaded .spec.yaml files should be saved")
	flag.StringVar(&region, "r", "us-east-1", "aws region where bucket is located")
	flag.Parse()

	if region == "" {
		fmt.Fprintf(os.Stderr, "error: S3 region must be specified\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if bucket == "" {
		fmt.Fprintf(os.Stderr, "error: S3 bucket must be specified\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if downloadDir == "" {
		fmt.Fprintf(os.Stderr, "error: download directory must be specified\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("parallelism =", parallelism)
	fmt.Println("bucket =", bucket)
	fmt.Println("prefix =", prefix)
	fmt.Println("download dir =", downloadDir)
}

func lower(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func downloadS3Objects(downloader *s3manager.Downloader, ns []string, done chan bool) {
	for _, n := range ns {
		p := fmt.Sprintf("%s/%s", downloadDir, path.Base(n))
		f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			errf("error: %v", err)
		}

		o := &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    aws.String(n),
		}
		if _, err = downloader.Download(f, o); err != nil {
			errf("error: %v", err)
		}
		f.Close()
	}
	done <- true
}

func errf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
