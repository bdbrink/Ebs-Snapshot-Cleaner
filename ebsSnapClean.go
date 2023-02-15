package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"
)

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
			aws.String("217906394988"),
		},
	}

	result, err := client.DescribeSnapshots(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	//fmt.Println(result)
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
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return false
	}

	//fmt.Println(result)
	return true
}

func seperateByDate(client ec2iface.EC2API) {

	oneMonth := time.Now().AddDate(0, -1, 0)
	swg := sizedwaitgroup.New(25)

	snapshots, err := getSnapshots(client)
	if err != nil {
		log.Errorf("Couldn't list snaps: %v", err)
	}

	for _, snapshot := range snapshots.Snapshots {
		//log.Infof("Found snapshot: %s created at %s \n", *snapshot.SnapshotId, *snapshot.StartTime)
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

}

func deleteSnapshot(client ec2iface.EC2API, snapshot string) {

	input := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(snapshot),
	}

	_, err := client.DeleteSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			case "InvalidSnapshot.InUse":
				log.Infof("skipping %s because snapshot is in use \n", snapshot)
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	log.Infof("%s has been deleted.", snapshot)
}
func main() {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	ec2Client := ec2.New(sess)
	seperateByDate(ec2Client)

	if err != nil {
		fmt.Println("error my guy")
	}

}
