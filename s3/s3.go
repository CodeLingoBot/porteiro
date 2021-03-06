// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package s3 provides an opener for S3 resources declared in the format
// s3://<bucket-name>/<object-key>.
package s3

import (
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3iface"
	"github.com/fsouza/porteiro"
)

type s3Opener struct {
	client s3iface.S3API
}

func newS3Opener(client s3iface.S3API) (*s3Opener, error) {
	if client == nil {
		cfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			return nil, err
		}
		client = s3.New(cfg)
	}
	return &s3Opener{client: client}, nil
}

func (o *s3Opener) open(url *url.URL) (io.ReadCloser, error) {
	req := o.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(url.Host),
		Key:    aws.String(strings.TrimLeft(url.Path, "/")),
	})
	object, err := req.Send()
	if err != nil {
		return nil, err
	}
	return object.Body, nil
}

// Open returns an opener that is able of loading files from S3 via the "s3"
// scheme.
//
// The s3 client might be set to nil, in which case a new client will be
// created using the default credential provider chain (see
// https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default
// for more details).
func Open(client s3iface.S3API, o *porteiro.Opener) (*porteiro.Opener, error) {
	opener, err := newS3Opener(client)
	if err != nil {
		return nil, err
	}
	return o.Register("s3", opener.open), nil
}
