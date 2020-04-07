package websocket

import (
	"context"
	"encoding/json"

	// "flag"
	"fmt"
	pb "ggstudy/asd/webservice/agent/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "path/filepath"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	// "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	address = os.Getenv("ADDRESS")
	uuid    = os.Getenv("UUID")
	cid     = os.Getenv("CID")
	dtuuid  = os.Getenv("DTUUID")
	dtcid   = os.Getenv("DTCID")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// HandshakeTimeout:  30,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSController struct {
	beego.Controller
	StopCh chan struct{}
}

// HandleTerminalSession is used to handle exec Request
func (h *WSController) HandleTerminalSession() {
	containerinfo := &ContainerInfo{
		Namespace: h.GetString("namespace"),
		Pod:       h.GetString("pod"),
		Container: h.GetString("container"),
	}
	fmt.Println(containerinfo)

	conn, err := upgrader.Upgrade(h.Ctx.ResponseWriter, h.Ctx.Request, nil)
	if err != nil {
		log.Printf("upgrader error: err=%v", err)
		http.Error(h.Ctx.ResponseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	terminalSession := TerminalSession{
		ContainerInfo: containerinfo,
		WSConn:        conn,
		Alive:         make(chan bool),
		Done:          make(chan struct{}),
		SizeChan:      make(chan remotecommand.TerminalSize),
	}

	defer terminalSession.Close()

	var (
		buf []byte
		msg TerminalMessage
	)

	if _, buf, err = conn.ReadMessage(); err != nil {
		log.Printf("HandleTerminalSession: can't Recv: %v", err)
		return
	}

	if err = json.Unmarshal(buf, &msg); err != nil {
		log.Printf("HandleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}

	if msg.Op != "bind" {
		log.Printf("HandleTerminalSession: expected 'bind' message, got: %s", msg.Op)
		return
	}

	go terminalSession.Ping(h.StopCh)

	dr := getPID()
	err = startProcess(&terminalSession, terminalSession, dr.Pid, dr.Dtpid)
	if err != nil {
		log.Printf("Error occured on remote connection: err=%v", err)
	}
}

// ExecCmd exec command on specific pod and wait the command's output.
func startProcess(t *TerminalSession, ptyHandler PtyHandler, pid, dtpid string) error {
	globalEndPoint, globalBearToken := getGlobalK8SAPIInfo()
	config := GenerateRestConfig(globalEndPoint, globalBearToken)
	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	if config.GroupVersion == nil {
		config.GroupVersion = &schema.GroupVersion{}
	}
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	}

	cs := kubernetes.NewForConfigOrDie(config)
	req := cs.CoreV1().RESTClient().Post().Resource("pods").Name(t.Pod).Namespace(t.Namespace).
		SubResource("exec").Param("container", t.Container)

	req.VersionedParams(&v1.PodExecOptions{
		Container: t.Container,
		Command:   []string{"./mnt/agent/letmein", pid, dtpid},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)
	fmt.Println("madebi")

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
}

func getPID() *pb.DebugResult {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb.NewDebugClient(conn)

	r, err := c.Debug(context.Background(),
		&pb.DebugRequest{
			TargetPodUUID:               uuid,
			TargetContainerID:           cid,
			TargetDebugToolsPodUUID:     dtuuid,
			TargetDebugToolsContainerID: dtcid,
		})
	if err != nil {
		log.Println(err)
	}
	return r
}

func GenerateRestConfig(ep, tk string) *rest.Config {
	cf := &rest.Config{
		Host:            ep,
		BearerToken:     tk,
		Timeout:         time.Duration(30) * time.Second,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	if cf.APIPath == "" {
		cf.APIPath = "/api"
	}
	if cf.GroupVersion == nil {
		cf.GroupVersion = &schema.GroupVersion{}
	}
	if cf.NegotiatedSerializer == nil {
		cf.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	}
	return cf
}

func getGlobalK8SAPIInfo() (endpoint, token string) {
	tk, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Fatal(err)
	}
	token = strings.Replace(string(tk), "\n", "", 1)
	endpoint = "https://10.96.0.1:443"
	return endpoint, token
}
