package main_test

import (
	"fmt"
	"testing"

	"github.com/ShaunPark/nfsMonitor/nfs"
	"github.com/ShaunPark/nfsMonitor/rest"
)

func TestMount(t *testing.T) {
	list := rest.GetNFSVolumes()

	for _, v := range list {
		volume := v
		fmt.Printf("%v ", v)
		if err := nfs.TestMountWithTimeout(volume.Host, volume.RemotePath, volume.SubPath, 5); err == nil {
			rest.UpdateVolume(volume, true)
		} else {
			fmt.Print(err)
			rest.UpdateVolume(volume, false)
		}
	}
}
