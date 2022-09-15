package remotestate

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Svc struct {
	s3Cli *s3.S3
}

var s3Svc = newS3Client()

func newS3Client() *S3Svc {
	ses := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &S3Svc{s3Cli: s3.New(ses)}
}

func GetObjectVersions(bucket string, name string) []*s3.ObjectVersion {
	vInput := s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(name),
	}
	versions, err := s3Svc.s3Cli.ListObjectVersions(&vInput)
	if err != nil {
		log.Fatal(err)
	}

	return versions.Versions

}

func ListBucketObjects(bucket string, name string, maxkeys *int64) *s3.ListObjectsV2Output {
	vInput := s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(name),
		MaxKeys: maxkeys,
	}
	objects, err := s3Svc.s3Cli.ListObjectsV2(&vInput)
	if err != nil {
		log.Fatal(err)
	}

	return objects
}

func GetBucketObjects(bucket string, name string, v *s3.ObjectVersion) *s3.GetObjectOutput {
	gInput := s3.GetObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(name),
		VersionId: v.VersionId,
		// Range:     aws.String("bytes=0-100"),
	}
	object, err := s3Svc.s3Cli.GetObject(&gInput)
	if err != nil {
		log.Fatal(err)
	}

	return object
}
