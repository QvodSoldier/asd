package websocket

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	pb "ggstudy/asd/webservice/agent/grpc"
	"ggstudy/asd/webservice/models"
	"ggstudy/asd/webservice/models/crds/debugtask"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/grpc"

	corev1 "k8s.io/api/core/v1"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ()

var (
	Store = models.NewStore()
	Cache = models.NewCache()
)

type wsTask struct {
	containerinfo ContainerInfo
	// debugPodName    string
	debugTask       *debugtask.DebugTask
	agentGRPCServer string
}

func (t *wsTask) getDebugRequest() (*pb.DebugRequest, error) {
	targetPod, err := Cache.Client.CoreV1().Pods(t.containerinfo.Namespace).Get(t.containerinfo.Pod, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	dr := &pb.DebugRequest{
		TargetPodUUID:     string(targetPod.UID),
		TargetContainerID: getPodCountainerID(t.containerinfo.Container, targetPod.Status),
	}

	dt, err := Store.ClientSet.Debug(namespace).Create(newDebugTask(t.containerinfo))
	if err != nil {
		return nil, err
	}
	t.debugTask = dt

	debugPod, err := t.waitDebugPodRuning()
	if err != nil {
		log.Println(err)
		if err := t.updateDebugTaskStatus("Failed"); err != nil {
			log.Printf("update status to failed error: %v", err)
		}

		return nil, err
	}
	// t.debugPodName = debugPod.Name
	// t.agentGRPCServer = debugPod.Status.PodIP + ":12580"
	t.agentGRPCServer = "192.168.99.101:30388"

	dr.TargetDebugToolsPodUUID = string(debugPod.UID)
	dr.TargetDebugToolsContainerID = getPodCountainerID("asd", debugPod.Status)

	return dr, nil
}

func (t *wsTask) waitDebugPodRuning() (*corev1.Pod, error) {
	taskName := t.debugTask.Name
	for i := 0; i < 20; i++ {
		pod, err := Cache.Client.CoreV1().Pods(namespace).Get("debug-pod-"+taskName, metav1.GetOptions{})
		if err != nil && !apiErr.IsNotFound(err) {
			log.Println(err)
			return nil, err
		}

		if isRunning(pod.Status) {
			return pod, nil
		}

		time.Sleep(time.Duration(3) * time.Second)
	}

	return nil, errors.New("create debug pod time out")
}

func isRunning(status corev1.PodStatus) bool {
	if status.Phase == "Running" {
		for _, v := range status.Conditions {
			if v.Type == "Ready" && v.Status == "True" {
				return true
			}
		}
	}

	return false
}

func getPodCountainerID(containerName string, status corev1.PodStatus) string {
	for _, v := range status.ContainerStatuses {
		if v.Name == containerName {
			tmp := strings.SplitN(v.ContainerID, "//", -1)
			return tmp[1]
		}
	}

	return ""
}

func (t *wsTask) getPID(dr *pb.DebugRequest) (*pb.DebugResult, error) {
	conn, err := grpc.Dial(t.agentGRPCServer, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewDebugClient(conn)

	r, err := c.Debug(context.Background(), dr)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (t *wsTask) updateDebugTaskStatus(status string) error {
	dt, err := Store.ClientSet.Debug(namespace).Get(t.debugTask.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	dt.Status.Phase = debugtask.DebugPhase(status)
	dt, err = Store.ClientSet.Debug(namespace).Update(dt)
	if err != nil {
		return err
	}

	t.debugTask = dt
	return nil
}

func newDebugTask(containerinfo ContainerInfo) *debugtask.DebugTask {
	tmp := &debugtask.DebugTask{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DebugTask",
			APIVersion: "debug.mahuang.cn/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "task-" + getUID(),
			// TODO: 改成指定的命名空间
			Namespace: namespace,
		},
		Spec: debugtask.DebugTaskSpec{
			DebugObjectInfo: &debugtask.DebugObjectInfo{
				DebugPodImage: containerinfo.DebugImage,
			},
			TargetObjectInfo: &debugtask.TargetObjectInfo{
				TargetPodNamespace:     containerinfo.Namespace,
				TargetPodName:          containerinfo.Pod,
				TargetPodContainerName: containerinfo.Container,
			},
		},
	}

	return tmp
}

func getUID() string {
	id := uuid.NewV4()

	tmp := strings.SplitN(id.String(), "-", -1)
	return tmp[0]
}
