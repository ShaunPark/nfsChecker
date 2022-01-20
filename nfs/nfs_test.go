package nfs_test

import (
	"log"
	"testing"

	"github.com/ShaunPark/nfsMonitor/nfs"
)

func TestPtrn(t *testing.T) {
	err := nfs.TestMountWithTimeout("10.177.235.105", "/data/nfstest", "mirrorswp", 5)
	if err != nil {
		log.Fatal(err)
	}
}
