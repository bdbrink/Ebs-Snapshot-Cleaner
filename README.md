# Ebs-Snapshot-Cleaner
Cleans old ebs snapshots

## How this works

This will scan all snapshots and AMIs in your aws account, if the snapshot does not have a corresponding AMI it will attempt to delete it.

## Throttling Issues

Due to the amount of snapshots and AMIs in accounts this has a tendenacy to run very long, using go concurrency speeds up the app, but due amount of API calls to EC2 it will get throttled (AWS thinks it could be nefarious), ex 1500 snapshots/AMIs, checking every snap and AMI 50 at a time with one second pause between go routines.