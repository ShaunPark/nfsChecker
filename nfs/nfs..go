package nfs

import (
	"fmt"
	"time"

	nfs "github.com/vmware/go-nfs-client/nfs"
	rpc "github.com/vmware/go-nfs-client/nfs/rpc"
	util "github.com/vmware/go-nfs-client/nfs/util"
)

func TestMountWithTimeout(host string, mountdir string, dir string, timeout int) error {
	to := make(chan bool, 1)
	ret := make(chan bool, 1)

	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		to <- true
	}()

	go func() {
		ret <- testMount(host, mountdir, dir)
	}()

	select {
	case <-to:
		return fmt.Errorf("timeout to mount check : %s:%s", host, mountdir)
	case isSuccess := <-ret:
		if isSuccess {
			return nil
		}
		return fmt.Errorf("failed to mount check : %s:%s", host, mountdir)
	}
}

func testMount(host string, mountdir string, dir string) bool {
	fmt.Printf("testMount started : %s %s %s", host, mountdir, dir)
	util.Infof("host=%s mountdir=%s dir=%s\n", host, mountdir, dir)

	mount, err := nfs.DialMount(host)
	if err != nil {
		// log.Fatalf("unable to dial MOUNT service: %v", err)
		fmt.Printf("%s", err.Error())
		return false
	}
	defer mount.Close()

	auth := rpc.NewAuthUnix("root", 1, 1)

	v, err := mount.Mount(mountdir, auth.Auth())
	if err != nil {
		fmt.Printf("%s", err.Error())
		return false
	}

	err = mount.Unmount()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return false
	}
	defer v.Close()
	fmt.Printf("testMount ended : %s %s %s", host, mountdir, dir)

	return true
}
