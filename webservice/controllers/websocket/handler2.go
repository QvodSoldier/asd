package websocket

import (
	"log"
	"net/http"
	"sync"

	"ggstudy/asd/webservice/models"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"errors"
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"
)

type TSockjs struct {
	beego.Controller
}

type ContainerInfo struct {
	DebugImage string `json:"debugImage"`
	Namespace  string `json:"namespace"`
	Pod        string `json:"pod"`
	Container  string `json:"container"`
}

var (
	// wsSocket *websocket.Conn
	// msg *WsMessage
	// copyData []byte
	// wsConn *WsConnection
	// sshReq *rest.Request
	// podName string
	// podNs string
	// container string
	// executor remotecommand.Executor
	// handler *streamHandler
	// err error
	namespace = viper.GetString("NAMESPACE")
	msgType   int
	data      []byte
)

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// websocket消息
type WsMessage struct {
	MessageType int
	Data        []byte
}

// 封装websocket连接
type WsConnection struct {
	wsSocket  *websocket.Conn // 底层websocket
	inChan    chan *WsMessage // 读取队列
	outChan   chan *WsMessage // 发送队列
	mutex     sync.Mutex      // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

// 读取协程
func (wsConn *WsConnection) wsReadLoop() {
	for {
		// 读一条message
		msgType, data, err := wsConn.wsSocket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// 放入请求队列
		wsConn.inChan <- &WsMessage{
			msgType,
			data,
		}

		//select {
		//case wsConn.inChan <- msg:
		//case <- wsConn.closeChan:
		//
		//}
	}
}

// 发送协程
func (wsConn *WsConnection) wsWriteLoop() {
	// 服务端返回给页面的数据

	for {
		select {
		// 取一个应答
		case msg := <-wsConn.outChan:
			// 写给web  websocket

			if err := wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				break
			}
		case <-wsConn.closeChan:
			wsConn.WsClose()
		}
	}

}

func InitWebsocket(resp http.ResponseWriter, req *http.Request) (wsConn *WsConnection, err error) {
	// 应答客户端告知升级连接为websocket
	wsSocket1, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return nil, err
	}

	wsConn = &WsConnection{
		wsSocket:  wsSocket1,
		inChan:    make(chan *WsMessage, 1000),
		outChan:   make(chan *WsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}

	// 页面读入输入 协程
	go wsConn.wsReadLoop()
	// 服务端返回数据 协程
	go wsConn.wsWriteLoop()

	return
}

// 发送返回消息到协程
func (wsConn *WsConnection) WsWrite(messageType int, data []byte) error {
	select {
	case wsConn.outChan <- &WsMessage{messageType, data}:

	case <-wsConn.closeChan:
		err := errors.New("WsWrite websocket closed")
		return err
	}
	return nil
}

// 读取 页面消息到协程
func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg := <-wsConn.inChan:
		return msg, err
	case <-wsConn.closeChan:
		err := errors.New("WsRead websocket closed")
		return nil, err
	}
}

// 关闭连接
func (wsConn *WsConnection) WsClose() {
	wsConn.wsSocket.Close()
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}

// ssh流式处理器
type streamHandler struct {
	wsConn      *WsConnection
	resizeEvent chan remotecommand.TerminalSize
}

// web终端发来的包
type xtermMessage struct {
	MsgType string `json:"type"`  // 类型:resize客户端调整终端, input客户端输入
	Input   string `json:"input"` // msgtype=input情况下使用
	Rows    uint16 `json:"rows"`  // msgtype=resize情况下使用
	Cols    uint16 `json:"cols"`  // msgtype=resize情况下使用
}

// executor回调获取web是否resize
func (handler *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}

// executor回调读取web端的输入
func (handler *streamHandler) Read(p []byte) (size int, err error) {

	// 读web发来的输入
	msg, err := handler.wsConn.WsRead()
	if err != nil {
		handler.wsConn.WsClose()
		return
	}
	// 解析客户端请求
	//if err = json.Unmarshal([]byte(msg.Data), &xtermMsg); err != nil {
	//	return
	//}

	xtermMsg := &xtermMessage{
		//MsgType: string(msg.MessageType),
		Input: string(msg.Data),
	}
	// 放到channel里，等remotecommand executor调用我们的Next取走
	handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	size = len(xtermMsg.Input)
	copy(p, xtermMsg.Input)
	return

}

// executor回调向web端输出
func (handler *streamHandler) Write(p []byte) (size int, err error) {
	// 产生副本
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.wsConn.WsWrite(websocket.TextMessage, copyData)
	return
}

func (t *TSockjs) ServeHTTP() {
	// t.EnableRender = false
	containerinfo := &ContainerInfo{
		DebugImage: t.GetString("debugImage"),
		Namespace:  t.GetString("namespace"),
		Pod:        t.GetString("pod"),
		Container:  t.GetString("container"),
	}
	fmt.Println(containerinfo)

	wsConn, err := InitWebsocket(t.Ctx.ResponseWriter, t.Ctx.Request)
	if err != nil {
		fmt.Println("wsConn err", err)
		wsConn.WsClose()
	}

	datas, _ := ioutil.ReadFile("conf/titletext")
	wsConn.WsWrite(websocket.TextMessage, datas)

	if err := startProcess(containerinfo, wsConn); err != nil {
		log.Println(err)
	}

	t.TplName = "terminal.html"
}

func startProcess(t *ContainerInfo, wsConn *WsConnection) error {
	wst := &wsTask{
		containerinfo: *t,
	}
	// 创建debugTask，等待debug容器启动，得到4个uuid和dt的地址,构造pb.DebugRequst,获取两个pid
	drq, err := wst.getDebugRequest()
	if err != nil {
		log.Printf("get debug request error: %v", err)
		return err
	}

	drs, err := wst.getPID(drq)
	if err != nil {
		if err := wst.updateDebugTaskStatus("Failed"); err != nil {
			log.Printf("update status to failed error: %v", err)
		}
		log.Printf("get pid error: %v", err)
		return err
	}

	if err = wst.updateDebugTaskStatus("Debuging"); err != nil {
		log.Printf("update status to debuging error: %v", err)
		return err
	}

	config := models.Cf
	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	if config.GroupVersion == nil {
		config.GroupVersion = &schema.GroupVersion{}
	}
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	}

	req := Cache.Client.CoreV1().RESTClient().Post().Resource("pods").
		Name(wst.debugTask.Spec.DebugObjectInfo.DebugPodName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", "asd")

	log.Println(drs.GetPid(), drs.GetDtpid())
	req.VersionedParams(&v1.PodExecOptions{
		Container: "asd",
		Command:   []string{"./mnt/agent/letmein", drs.GetPid(), drs.GetDtpid()},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	// 创建到容器的连接
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		wsConn.WsClose()
		return err
	}

	// 配置与容器之间的数据流处理回调
	handler := &streamHandler{wsConn: wsConn, resizeEvent: make(chan remotecommand.TerminalSize)}
	if err := executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	}); err != nil {
		fmt.Println("handler", err)
		return err

	}
	return err

}
