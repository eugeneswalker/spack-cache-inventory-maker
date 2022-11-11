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
	"path/filepath"
	"io"
	"bytes"
	"bufio"
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
			if strings.HasSuffix(k, ".spec.json") || strings.HasSuffix(k, ".spec.yaml") || strings.HasSuffix(k, ".spec.json.sig") {
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

func lineCounter(r io.Reader) (int, error) {
    buf := make([]byte, 32*1024)
    count := 0
    lineSep := []byte{'\n'}

    for {
        c, err := r.Read(buf)
        count += bytes.Count(buf[:c], lineSep)

        switch {
        case err == io.EOF:
            return count, nil

        case err != nil:
            return count, err
        }
    }
}

func downloadS3Objects(downloader *s3manager.Downloader, ns []string, done chan bool) {
	for _, n := range ns {
		p := fmt.Sprintf("%s/%s", downloadDir, path.Base(n))

		if _, err := os.Stat(p); err == nil {
			// already downloaded
			fmt.Println("Already downloaded", p)
			continue
		}

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

		if strings.HasSuffix(n, ".spec.json.sig") {
			// open .spec.json.sig
			sigf, err := os.Open(p)
			if err != nil {
				errf("error: %v", err)
			}

			nLines, err := lineCounter(sigf)
			if err != nil {
				errf("error: %v", err)
			}

			// create corresponding .spec.json
			jsonFn := strings.TrimSuffix(p, filepath.Ext(p))
			jsonf, err := os.OpenFile(jsonFn, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				errf("error: %v", err)
			}

			// copy only json contents of .spec.json.sig file to .spec.json file
			sigf.Seek(0, io.SeekStart)
			scanner := bufio.NewScanner(sigf)
			scanner.Split(bufio.ScanLines)
			i := 0
			for scanner.Scan() {
				if i < 3 {
					i++
					continue
				}
				fmt.Fprintln(jsonf, scanner.Text())
				i++
				if i >= nLines - 16 {
					break
				}
			}
			sigf.Close()
			jsonf.Close()

			err = os.Rename(jsonFn, p)
			if err != nil {
				errf("error: %v", err)
			}
		}
	}
	done <- true
}

func errf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
