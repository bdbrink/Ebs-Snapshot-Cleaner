package main

import (
	"testing"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
)

// mock ec2 client
type mockEc2Client struct {
	ec2iface.EC2API
}

func (m *mockEc2Client) DescribeSnapshots(*ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error) {

	snapshot1 := ec2.Snapshot{
		SnapshotId: aws.String("test1234"),
		StartTime:  aws.Time(time.Now()),
	}

	snapshot2 := ec2.Snapshot{
		SnapshotId: aws.String("testDeleteMe"),
		StartTime:  aws.Time(time.Now().AddDate(0, -4, 0)),
	}

	output := &ec2.DescribeSnapshotsOutput{
		Snapshots: []*ec2.Snapshot{
			&snapshot1, &snapshot2,
		},
	}

	return output, nil
}

func (m *mockEc2Client) DeleteSnapshot(*ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error) {

	output := &ec2.DeleteSnapshotOutput{}

	return output, nil
}

// return 2 snaps defined in mock api
func TestGetSnapshots(t *testing.T) {

	mockClient := &mockEc2Client{}
	testGetSnaps, _ := getSnapshots(mockClient)
	t.Run("return mock snapshots", func(t *testing.T) {
		assert.Equal(t, testGetSnaps, testGetSnaps)
	})

}

// snap testDeleteMe will be deleted
func TestSeperateByDate(t *testing.T) {

	mockClient := &mockEc2Client{}
	testSeperateByDate := seperateByDate(mockClient)
	t.Run("seperate snaps and delete by date", func(t *testing.T) {
		assert.Equal(t, testSeperateByDate, testSeperateByDate)
	})
}

// feed string to delete that snapshot
func TestDeleteSnapshots(t *testing.T) {

	mockClient := &mockEc2Client{}
	snapshot := "test1"
	testDelete := deleteSnapshot(mockClient, snapshot)

	t.Run("test snapshot deletion", func(t *testing.T) {
		assert.Equal(t, testDelete, testDelete)
	})
}
