package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

const bucket = "fieldservice-nightly"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("nightly snapshot uploaded and cleaned up")
}

func run() (err error) {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.createBucket(bucket); err != nil {
		return err
	}
	now := time.Now()
	key := snapshotKey(now)
	defer func() {
		err = errors.Join(err, c.deleteObject(bucket, key), c.deleteBucket(bucket))
	}()
	orders := []WorkOrder{{ID: "WO-1042", DispatchStatus: "complete", Photos: []string{"photos/WO-1042-before.jpg"}, FollowUp: "confirm customer sign-off"}}
	if err := uploadSnapshot(c, bucket, now, orders, httpPut); err != nil {
		return err
	}
	return nil
}
