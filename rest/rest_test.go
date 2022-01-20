package rest_test

import (
	"fmt"
	"testing"

	"github.com/ShaunPark/nfsMonitor/rest"
)

func TestAPI(t *testing.T) {
	volumes := rest.GetNFSVolumes()
	for _, v := range volumes {
		fmt.Printf("%+v\n", v)
	}
}

func TestUpdate(t *testing.T) {
	volumes := rest.GetNFSVolumes()
	for _, v := range volumes {
		rest.UpdateVolume(v, false)
	}
}
