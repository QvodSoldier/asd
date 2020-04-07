package server

import (
	"context"
	"errors"
	pb "ggstudy/asd/webservice/agent/grpc"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

const dir string = "/proc/"

type Server struct {
	savedDebugRequests []*pb.DebugRequest
}

func (s *Server) Debug(ctx context.Context, in *pb.DebugRequest) (*pb.DebugResult, error) {
	s.savedDebugRequests = append(s.savedDebugRequests, in)
	pid, err := getPid(in.GetTargetPodUUID(), in.GetTargetContainerID())
	dtpid, err := getPid(in.GetTargetDebugToolsPodUUID(), in.GetTargetDebugToolsContainerID())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(pid, dtpid)
	return &pb.DebugResult{
		Pid:   pid,
		Dtpid: dtpid,
	}, nil
}

func getPid(podUUID, contianerID string) (string, error) {
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if matched, _ := regexp.MatchString("^(\\d+)$", f.Name()); matched {
			content, _ := ioutil.ReadFile(dir + f.Name() + "/cgroup")
			if strings.Contains(string(content), podUUID) &&
				strings.Contains(string(content), contianerID) {
				return f.Name(), nil
			}
		}
	}
	return "", errors.New("can't find this pod's pid")
}
