package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ShaunPark/nfsMonitor/types"
)

func GetNFSVolumes() []*types.VOLUME {
	apiServer := "localhost:9099"

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/api/v1/projects/volumes", apiServer), nil)
	if err != nil {
		panic(err)
	}

	//필요시 헤더 추가 가능
	req.Header.Add("Accept", "application/json")
	req.Header.Add("implicit-authenticated-for", "bee.admin")
	req.Header.Add("Authorization", "Bearer implicit-token")

	q := req.URL.Query()
	q.Add("volume_type", "external")
	req.URL.RawQuery = q.Encode()

	// Client객체에서 Request 실행
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 결과 출력
	bytes, _ := ioutil.ReadAll(resp.Body)
	res := types.VOLUME_RESPONSE{}
	if err := json.Unmarshal(bytes, &res); err == nil {
		return res.Data
	}
	return nil
}

func UpdateVolume(v *types.VOLUME, s bool) {
	apiServer := "localhost:9099"

	volume := types.VOLUME_STATUS_UPDATE_REQUEST{Status: s}
	pbytes, _ := json.Marshal(volume)
	buff := bytes.NewBuffer(pbytes)

	// Request 객체 생성
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/v2/volumes/%d/status", apiServer, v.Id), buff)
	if err != nil {
		panic(err)
	}

	//Content-Type 헤더 추가
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("implicit-authenticated-for", "bee.admin")
	req.Header.Add("Authorization", "Bearer implicit-token")

	// Client객체에서 Request 실행
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		str := string(respBody)
		println(str)
	}
}
