package websocket

import (
	"os"
	"os/signal"
	"syscall"
)

// HandleSession validate if the session exists with the given id
// It delete the session after get
// func HandleSession(t *TerminalSession, sessionID string) error {
//
// 	redis := t.Redis
//
// 	key := fmt.Sprintf("ws-session-id:%s", sessionID)
//
// 	val, err := redis.Get(key)
//
// 	if err != nil {
// 		t.Logger.Errorf("Can't get key from redis: key=%s err=%v", key, err)
// 		return err
// 	}
//
// 	t.Logger.Debugf("Get session: val=%v", val)
//
// 	var containerInfo ContainerInfo
//
// 	err = json.Unmarshal([]byte(val), &containerInfo)
//
// 	if err != nil {
// 		t.Logger.Errorf("Can't Unmarshal: id=%s val=%v err=%v", sessionID, val, err)
// 		return err
// 	}
//
// 	t.ContainerInfo = &containerInfo
//
// 	return nil
// }

var onlyOneSignalHandler = make(chan struct{})

// SetupSignalHandler ..
func SetupSignalHandler() chan struct{} {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
	}()

	return stop
}
