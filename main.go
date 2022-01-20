package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ShaunPark/nfsMonitor/kubernetes"
	"github.com/ShaunPark/nfsMonitor/nfs"
	"github.com/ShaunPark/nfsMonitor/processor"
	"github.com/ShaunPark/nfsMonitor/rabbitmq"
	"github.com/ShaunPark/nfsMonitor/rest"
	"github.com/ShaunPark/nfsMonitor/types"
	"github.com/ShaunPark/nfsMonitor/utils"

	"gopkg.in/alecthomas/kingpin.v2"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

var (
	kubecfg                     = kingpin.Flag("kubeconfig", "Path to kubeconfig file. Leave unset to use in-cluster config.").String()
	apiserver                   = kingpin.Flag("master", "Address of Kubernetes API server. Leave unset to use in-cluster config.").String()
	namespace                   = kingpin.Flag("namespace", "Namespace used to create leader election lock object.").Short('n').Default("monitoring").String()
	leaderElectionTokenName     = kingpin.Flag("leader-election-token-name", "Leader election token name.").Default(component).String()
	leaderElectionLeaseDuration = kingpin.Flag("leader-election-lease-duration", "Lease duration for leader election.").Default(DefaultLeaderElectionLeaseDuration.String()).Duration()
	leaderElectionRenewDeadline = kingpin.Flag("leader-election-renew-deadline", "Leader election renew deadline.").Default(DefaultLeaderElectionRenewDeadline.String()).Duration()
	leaderElectionRetryPeriod   = kingpin.Flag("leader-election-retry-period", "Leader election retry period.").Default(DefaultLeaderElectionRetryPeriod.String()).Duration()

	configFile = kingpin.Flag("configFile", "configFile").Short('f').Required().String()
	logLevel   = kingpin.Flag("verbose", "log level").Short('v').Default("3").String()
	// dryRun     = kingpin.Flag("dryRun", "go ").Bool()

	responseCh chan *types.BeeResponse
)

const (
	DefaultLeaderElectionLeaseDuration time.Duration = 15 * time.Second
	DefaultLeaderElectionRenewDeadline time.Duration = 10 * time.Second
	DefaultLeaderElectionRetryPeriod   time.Duration = 2 * time.Second

	component = "nfs-server-mon"
)

func main() {
	kingpin.Parse()
	klog.InitFlags(nil)
	flag.Lookup("v").Value.Set(*logLevel)

	k := kubernetes.NewClient(apiserver, kubecfg)
	// cmg := utils.NewConfigManager(*configFile)

	stopCh := make(chan interface{})

	// elector code
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := NfsMon{
		configMgr: utils.NewConfigManager(*configFile),
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		defer close(stopCh)
		cancel()
	}()

	id, _ := os.Hostname()

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      *leaderElectionTokenName,
			Namespace: *namespace,
		},
		Client: k.Clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   *leaderElectionLeaseDuration,
		RenewDeadline:   *leaderElectionRenewDeadline,
		RetryPeriod:     *leaderElectionRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				go m.Run(stopCh)
			},
			OnStoppedLeading: func() {
				kingpin.Fatalf("lost leader election")
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == id {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
}

type NfsMon struct {
	configMgr *utils.ConfigManager
}

func (m NfsMon) Run(stopCh chan interface{}) {
	config := m.configMgr.GetConfig()
	interval := config.ProcessInterval

	klog.Infof("Start node memory montoring. Duration [%d]", (interval / 1000))

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-stopCh
		ticker.Stop()
		klog.Info("Stopping NodeManager...")
		wg.Done()
	}()

	// make channel for response to Rabbit MQ
	responseCh = make(chan *types.BeeResponse)

	handler := processor.NewHandler(responseCh, stopCh)

	rmqClient, err := rabbitmq.NewRabbitMQClient(config.RabbitMQ, handler.Proc)
	if err != nil {
		fmt.Print(err.Error())
		close(stopCh)
	} else {
		go func() {
			for {
				select {
				case value := <-responseCh:
					replyQueue := value.MetaData.Queue
					klog.Info("reply queue from metadata ", replyQueue)
					if !config.RabbitMQ.UseReplyQueueFromMessage || replyQueue == "" {
						replyQueue = config.RabbitMQ.ReplyQueue
					}
					bytes, _ := json.Marshal(*value)

					rmqClient.SendResponse(&replyQueue, &value.MetaData.CorrelationId, bytes)
				case <-stopCh:
					klog.Info("End of responseCh select")
					return
				}
			}
		}()
		go func() {
			klog.Info(" [*] Waiting for messages. To exit press CTRL+C")
			<-stopCh
			// proc.Close()
			rmqClient.Close()
		}()
		rmqClient.Run(stopCh)
	}

	go func() {
		for range ticker.C {
			checkAndUpdate()
		}
	}()
	wg.Wait()
	klog.Info("End of NFS Monitor...")
}

func checkAndUpdate() {
	list := rest.GetNFSVolumes()

	for _, v := range list {
		volume := v
		checkVolumeStatus(volume)
	}
}

func checkVolumeStatus(volume *types.VOLUME) {
	if err := nfs.TestMountWithTimeout(volume.Host, volume.RemotePath, volume.SubPath, 5); err == nil {
		rest.UpdateVolume(volume, true)
	} else {
		rest.UpdateVolume(volume, false)
	}
}

func NewEventRecorder(c k8s.Interface) record.EventRecorder {
	b := record.NewBroadcaster()
	b.StartRecordingToSink(&typedcore.EventSinkImpl{Interface: typedcore.New(c.CoreV1().RESTClient()).Events("")})
	return b.NewRecorder(scheme.Scheme, core.EventSource{Component: component})
}
