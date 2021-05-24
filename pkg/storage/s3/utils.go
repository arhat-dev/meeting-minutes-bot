/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2020 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3

import (
	"context"
	"net"
	"net/url"

	"github.com/minio/minio-go/v7/pkg/s3utils"
)

func (s *S3) formatPublicURL(ctx context.Context, objectKey string) string {
	location := s.region
	if location == "" {
		if len(s.bucket) != 0 {
			// Gather location only if bucketName is present.
			location, _ = s.getBucketLocation(ctx)
		}

		if location == "" {
			location = getDefaultLocation(*s.client.EndpointURL(), s.region)
		}
	}

	// Look if target url supports virtual host.
	// We explicitly disallow MakeBucket calls to not use virtual DNS style,
	// since the resolution may fail.
	isMakeBucket := false
	isVirtualHost := s.isVirtualHostStyleRequest(*s.client.EndpointURL(), s.bucket) && !isMakeBucket

	// Construct a new target URL.
	return s.makeTargetURL(s.bucket, objectKey, location, isVirtualHost)
}

// Get default location returns the location based on the input
// URL `u`, if region override is provided then all location
// defaults to regionOverride.
//
// If no other cases match then the location is set to `us-east-1`
// as a last resort.
func getDefaultLocation(u url.URL, regionOverride string) (location string) {
	if regionOverride != "" {
		return regionOverride
	}
	region := s3utils.GetRegionFromURL(u)
	if len(region) == 0 {
		region = "us-east-1"
	}
	return region
}

// getBucketLocation - Get location for the bucketName from location map cache, if not
// fetch freshly by making a new request.
func (s *S3) getBucketLocation(ctx context.Context) (string, error) {
	// Region set then no need to fetch bucket location.
	if len(s.region) != 0 {
		return s.region, nil
	}

	return s.client.GetBucketLocation(ctx, s.bucket)
}

// returns true if virtual hosted style requests are to be used.
func (s *S3) isVirtualHostStyleRequest(url url.URL, bucketName string) bool {
	if len(bucketName) == 0 {
		return false
	}

	// default to virtual only for Amazon/Google  storage. In all other cases use
	// path style requests
	return s3utils.IsVirtualHostSupported(url, bucketName)
}

// makeTargetURL make a new target url.
func (s *S3) makeTargetURL(bucketName, objectName, bucketLocation string, isVirtualHostStyle bool) string {
	host := s.client.EndpointURL().Host
	// For Amazon S3 endpoint, try to fetch location based endpoint.
	if s3utils.IsAmazonEndpoint(*s.client.EndpointURL()) {
		// if c.s3AccelerateEndpoint != "" && bucketName != "" {
		// 	// http://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html
		// 	// Disable transfer acceleration for non-compliant bucket names.
		// 	if strings.Contains(bucketName, ".") {
		// 		return nil, errTransferAccelerationBucket(bucketName)
		// 	}
		// 	// If transfer acceleration is requested set new host.
		// 	// For more details about enabling transfer acceleration read here.
		// 	// http://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html
		// 	host = c.s3AccelerateEndpoint
		// } else {

		// }

		// Do not change the host if the endpoint URL is a FIPS S3 endpoint.
		if !s3utils.IsAmazonFIPSEndpoint(*s.client.EndpointURL()) {
			// Fetch new host based on the bucket location.
			host = getS3Endpoint(bucketLocation)
		}
	}

	// Save scheme.
	scheme := s.client.EndpointURL().Scheme

	// Strip port 80 and 443 so we won't send these ports in Host header.
	// The reason is that browsers and curl automatically remove :80 and :443
	// with the generated presigned urls, then a signature mismatch error.
	if h, p, err := net.SplitHostPort(host); err == nil {
		if scheme == "http" && p == "80" || scheme == "https" && p == "443" {
			host = h
		}
	}

	urlStr := scheme + "://" + host + "/"
	// Make URL only if bucketName is available, otherwise use the
	// endpoint URL.
	if bucketName != "" {
		// If endpoint supports virtual host style use that always.
		// Currently only S3 and Google Cloud Storage would support
		// virtual host style.
		if isVirtualHostStyle {
			urlStr = scheme + "://" + bucketName + "." + host + "/"
			if objectName != "" {
				urlStr += s3utils.EncodePath(objectName)
			}
		} else {
			// If not fall back to using path style.
			urlStr += bucketName + "/"
			if objectName != "" {
				urlStr += s3utils.EncodePath(objectName)
			}
		}
	}

	return urlStr
}

// awsS3EndpointMap Amazon S3 endpoint map.
var awsS3EndpointMap = map[string]string{
	"us-east-1":      "s3.dualstack.us-east-1.amazonaws.com",
	"us-east-2":      "s3.dualstack.us-east-2.amazonaws.com",
	"us-west-2":      "s3.dualstack.us-west-2.amazonaws.com",
	"us-west-1":      "s3.dualstack.us-west-1.amazonaws.com",
	"ca-central-1":   "s3.dualstack.ca-central-1.amazonaws.com",
	"eu-west-1":      "s3.dualstack.eu-west-1.amazonaws.com",
	"eu-west-2":      "s3.dualstack.eu-west-2.amazonaws.com",
	"eu-west-3":      "s3.dualstack.eu-west-3.amazonaws.com",
	"eu-central-1":   "s3.dualstack.eu-central-1.amazonaws.com",
	"eu-north-1":     "s3.dualstack.eu-north-1.amazonaws.com",
	"eu-south-1":     "s3.dualstack.eu-south-1.amazonaws.com",
	"ap-east-1":      "s3.dualstack.ap-east-1.amazonaws.com",
	"ap-south-1":     "s3.dualstack.ap-south-1.amazonaws.com",
	"ap-southeast-1": "s3.dualstack.ap-southeast-1.amazonaws.com",
	"ap-southeast-2": "s3.dualstack.ap-southeast-2.amazonaws.com",
	"ap-northeast-1": "s3.dualstack.ap-northeast-1.amazonaws.com",
	"ap-northeast-2": "s3.dualstack.ap-northeast-2.amazonaws.com",
	"ap-northeast-3": "s3.dualstack.ap-northeast-3.amazonaws.com",
	"af-south-1":     "s3.dualstack.af-south-1.amazonaws.com",
	"me-south-1":     "s3.dualstack.me-south-1.amazonaws.com",
	"sa-east-1":      "s3.dualstack.sa-east-1.amazonaws.com",
	"us-gov-west-1":  "s3.dualstack.us-gov-west-1.amazonaws.com",
	"us-gov-east-1":  "s3.dualstack.us-gov-east-1.amazonaws.com",
	"cn-north-1":     "s3.dualstack.cn-north-1.amazonaws.com.cn",
	"cn-northwest-1": "s3.dualstack.cn-northwest-1.amazonaws.com.cn",
}

// getS3Endpoint get Amazon S3 endpoint based on the bucket location.
func getS3Endpoint(bucketLocation string) (s3Endpoint string) {
	s3Endpoint, ok := awsS3EndpointMap[bucketLocation]
	if !ok {
		// Default to 's3.dualstack.us-east-1.amazonaws.com' endpoint.
		s3Endpoint = "s3.dualstack.us-east-1.amazonaws.com"
	}
	return s3Endpoint
}
