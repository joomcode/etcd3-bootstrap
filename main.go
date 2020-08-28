package main

import (
	"log"
	"flag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	useEBS                    bool
	ebsVolumeName             string
	mountPoint                string
	blockDevice               string
	awsRegion                 string
	fileSystemFormatType      string
	fileSystemFormatArguments string
	domain                    string
	zoneId                    string
	TTL                       int64
)

func init() {
	flag.StringVar(&awsRegion, "aws-region", "eu-west-1", "AWS region this instance is on")
	flag.StringVar(&ebsVolumeName, "ebs-volume-name", "", "EBS volume to attach to this node")
	flag.StringVar(&mountPoint, "mount-point", "/var/lib/etcd", "EBS volume mount point")
	flag.StringVar(&blockDevice, "block-device", "/dev/nvme1n1", "Block device to attach as")
	flag.StringVar(&fileSystemFormatType, "filesystem-type", "ext4", "Linux filesystem format type")
	flag.StringVar(&fileSystemFormatArguments, "filesystem-arguments", "", "Linux filesystem format arguments")
	flag.BoolVar(&useEBS, "use-ebs", true, "Use EBS instead of instance store")
	flag.StringVar(&domain, "domain", "", "domain name")
	flag.StringVar(&zoneId, "zone-id", "", "AWS Zone Id for domain")
	flag.Int64Var(&TTL, "ttl", int64(60), "TTL for DNS Cache")
	flag.Parse()
}

func main() {
	// Initialize AWS session
	awsSession := session.Must(session.NewSession())

	// Create route53 client
	route53Svc := route53.New(awsSession)

	// Create ec2 and metadata svc clients with specified region
	ec2SVC := ec2.New(awsSession, aws.NewConfig().WithRegion(awsRegion))
	metadataSVC := ec2metadata.New(awsSession, aws.NewConfig().WithRegion(awsRegion))

	// obtain current AZ, required for finding volume
	availabilityZone, err := metadataSVC.GetMetadata("placement/availability-zone")
	if err != nil {
		panic(err)
	}

	if useEBS {
		volume, err := volumeFromName(ec2SVC, ebsVolumeName, availabilityZone)
		if err != nil {
			panic(err)
		}

		instanceID, err := metadataSVC.GetMetadata("instance-id")
		if err != nil {
			panic(err)
		}

		err = attachVolume(ec2SVC, instanceID, volume)
		if err != nil {
			panic(err)
		}
	}

	if err := ensureVolumeInited(blockDevice, fileSystemFormatType, fileSystemFormatArguments); err != nil {
		panic(err)
	}

	if err := ensureVolumeMounted(blockDevice, mountPoint); err != nil {
		panic(err)
	}

	if err := ensureVolumeWriteable(mountPoint); err != nil {
		panic(err)
	}

	idd, err := metadataSVC.GetInstanceIdentityDocument()
	if err != nil {
		panic(err)
	}

	if domain != "" && zoneId != "" {
		log.Printf("will update domain %s (zoneId: %s)", domain, zoneId)
		if err := ensureDNSRecord(route53Svc, idd.PrivateIP); err != nil {
			panic(err)
		}
	}
}
