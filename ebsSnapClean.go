package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"
)

// pass in account ID to list all snaps belonging to that account
func getSnapshots(client ec2iface.EC2API) (*ec2.DescribeSnapshotsOutput, error) {

	input := &ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("status"),
				Values: []*string{
					aws.String("completed"),
				},
			},
		},
		OwnerIds: []*string{
			aws.String("Insert account ID"),
		},
	}

	result, err := client.DescribeSnapshots(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Errorln(aerr.Error())
			}
		} else {
			log.Errorln(err.Error())
		}
	}
	return result, nil
}

func getAmi(client ec2iface.EC2API, snapshot string) bool {

	input := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("block-device-mapping.snapshot-id"),
				Values: []*string{
					aws.String(snapshot),
				},
			},
		},
	}

	_, err := client.DescribeImages(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Errorln(aerr.Error())
			}
		} else {
			log.Errorln(err.Error())
		}
		return false
	}

	return true
}

func seperateByDate(client ec2iface.EC2API) bool {

	oneMonth := time.Now().AddDate(0, -1, 0)
	swg := sizedwaitgroup.New(25)

	snapshots, err := getSnapshots(client)
	if err != nil {
		log.Errorf("Couldn't list snaps: %v", err)
		return false
	}

	for _, snapshot := range snapshots.Snapshots {
		log.Infof("Found snapshot: %s created at %s \n", *snapshot.SnapshotId, *snapshot.StartTime)
		snapshotDate := *snapshot.StartTime
		if snapshotDate.Before(oneMonth) {
			snapshotID := *snapshot.SnapshotId
			go func() {
				defer swg.Done()
				amiExists := getAmi(client, snapshotID)
				if amiExists == false {
					log.Infof("Deleting: %s \n", *snapshot.SnapshotId)
					deleteSnapshot(client, snapshotID)
				}
			}()
			swg.Add()
		}
	}
	swg.Wait()
	return true

}

func deleteSnapshot(client ec2iface.EC2API, snapshot string) bool {

	input := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(snapshot),
	}

	_, err := client.DeleteSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Errorln(aerr.Error())
			case "InvalidSnapshot.InUse":
				log.Infof("skipping %s because snapshot is in use \n", snapshot)
			}
		} else {
			log.Errorln(err.Error())
		}
		return false
	}

	log.Infof("%s has been deleted.", snapshot)
	return true
}
func main() {

	// Lot of snaps being deleted, retry when throttled
	myRetryer := client.DefaultRetryer{
		NumMaxRetries:    10,
		MaxThrottleDelay: 10 * time.Second,
	}

	sess, err := session.NewSession(&aws.Config{
		Region:  aws.String("us-east-1"),
		Retryer: myRetryer,
	})
	ec2Client := ec2.New(sess)
	seperateByDate(ec2Client)

	if err != nil {
		log.Errorln(err)
	}

}
