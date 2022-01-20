package processor

import (
	"encoding/json"

	"github.com/ShaunPark/nfsMonitor/errors"
	"github.com/ShaunPark/nfsMonitor/nfs"
	"github.com/ShaunPark/nfsMonitor/rest"
	"github.com/ShaunPark/nfsMonitor/types"
	"github.com/ShaunPark/nfsMonitor/utils"
	"github.com/streadway/amqp"
	klog "k8s.io/klog/v2"
)

type ReqHandler struct {
	responseChan chan *types.BeeResponse
	stopCh       chan interface{}
}

func NewHandler(rChan chan *types.BeeResponse, sChan chan interface{}) ReqHandler {
	return ReqHandler{responseChan: rChan, stopCh: sChan}
}

func (r ReqHandler) Proc(values <-chan amqp.Delivery) {
	klog.Info("Proc in ReqHander")
	for {
		select {
		case value, ok := <-values:
			if !ok {
				klog.Info("Message not ok")
				return
			}
			klog.Info("Message Received")

			req := &types.BeeRequest{}
			if err := json.Unmarshal(value.Body, req); err != nil {
				r.responseChan <- makeErrorResponse(value.CorrelationId, value.ReplyTo, string(value.Body))
			} else {
				klog.Info(string(value.Body))
				volume := req.PayLoad.Data
				checkVolumeStatus(volume)

				klog.Infof("Message Send to : %s", value.ReplyTo)

				res := &types.BeeResponse{
					MetaData: &types.MetaData{
						Type:          utils.MESSAGE_TYPE,
						From:          utils.MESSAGE_DEFAULT_FROM,
						To:            utils.MESSAGE_DEFAULT_TO,
						CorrelationId: value.CorrelationId,
						Queue:         value.ReplyTo,
					},
					PayLoad: types.ResponsePayLoad{
						Status:      utils.MSG_STATUS_SUCCESS,
						Message:     "success",
						ErrorDetail: "success",
						ErrorOrigin: "STORAGE_MONITOR",
					},
				}
				r.responseChan <- res
			}
			value.Ack(false)
		case <-r.stopCh:
			return
		}
	}
}

func checkVolumeStatus(volume *types.VOLUME) {
	if err := nfs.TestMountWithTimeout(volume.Host, volume.RemotePath, volume.SubPath, 5); err == nil {
		rest.UpdateVolume(volume, true)
	} else {
		rest.UpdateVolume(volume, false)
	}
}

func makeErrorResponse(coId, replyTo, body string) *types.BeeResponse {
	return &types.BeeResponse{
		MetaData: &types.MetaData{
			Type:          utils.MESSAGE_TYPE,
			From:          utils.MESSAGE_DEFAULT_FROM,
			To:            utils.MESSAGE_DEFAULT_TO,
			CorrelationId: coId,
			Queue:         replyTo,
		},
		PayLoad: types.ResponsePayLoad{
			Status:      utils.MSG_STATUS_ERROR,
			Code:        errors.SERVICE_INTERNAL,
			Message:     errors.SERVICE_INTERNAL_MSG,
			ErrorDetail: body,
			ErrorOrigin: utils.HA_PROXY,
		},
	}
}

// type procData struct {
// 	CoId   string `json:"coid"`
// 	ReplyQ string `json:"reply"`
// 	Data   []byte `json:"data"`
// }
