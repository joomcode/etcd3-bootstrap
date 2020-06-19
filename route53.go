package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func ensureDNSRecord(svc *route53.Route53, target string) error {
	log.Printf("will update dns record %s to point to %s", domain, target)

	if domain == "" || target == "" || zoneId == "" {
		return fmt.Errorf("Incomplete arguments: domain: %s, target: %s, zone-id: %s\n", domain, target, zoneId)
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("UPSERT"), // Required
					ResourceRecordSet: &route53.ResourceRecordSet{ // Required
						Name: aws.String(domain), // Required
						Type: aws.String("A"),    // Required
						ResourceRecords: []*route53.ResourceRecord{
							{ // Required
								Value: aws.String(target), // Required
							},
						},
						TTL: aws.Int64(TTL),
					},
				},
			},
			Comment: aws.String("updating etcd record for a new node"),
		},
		HostedZoneId: aws.String(zoneId), // Required
	}

	_, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}

	return nil
}
