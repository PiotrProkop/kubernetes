package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
)

const (
	eventDispatcherAddress = "localhost:5433"
	eventHandlerAddress    = "localhost:5444"
	name                   = "iso"
)

type EventHandler interface {
	Start(socketAddress string)
}

type eventHandler struct {
}

func (e *eventHandler) Start(socketAddress string) {
	glog.Infof("Starting eventHandler")
	lis, err := net.Listen("tcp", socketAddress)
	if err != nil {
		glog.Fatalf("failed to bind to socket address: %v", err)
	}
	s := grpc.NewServer()
	lifecycle.RegisterEventHandlerServer(s, e)
	if err := s.Serve(lis); err != nil {
		glog.Fatalf("failed to start event handler server: %v", err)
	}
}
func (e *eventHandler) Notify(context context.Context, event *lifecycle.Event) (reply *lifecycle.EventReply, err error) {
	if event.Kind == 0 {
		glog.Infof("Received PreStop event with such payload: %v\n", event.CgroupInfo)
		path := fmt.Sprintf("%s%s%s", "/sys/fs/cgroup/cpuset", event.CgroupInfo.Path, "/cpuset.cpus")
		glog.Infof("Our path: %v", path)
		err := ioutil.WriteFile(path, []byte("0-1"), 0644)
		if err != nil {
			glog.Fatalf("OOOOPS: %v", err)
		}

	} else {
		return nil, fmt.Errorf("Wrong event type")
	}
	return &lifecycle.EventReply{
		Error:      "",
		CgroupInfo: event.CgroupInfo,
	}, nil

}

func main() {
	flag.Parse()
	glog.Info("Staring ...")
	ctx := context.Background()
	cxn, err := grpc.Dial(eventDispatcherAddress, grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("failed to connect to eventDispatcher: %v", err)
	}
	client := lifecycle.NewEventDispatcherClient(cxn)
	glog.Infof("Registering handler: %s\n", name)
	registerToken := string(uuid.NewUUID())
	registerRequest := &lifecycle.RegisterRequest{
		SocketAddress: eventHandlerAddress,
		Name:          name,
		Token:         registerToken,
	}
	reply, err := client.Register(ctx, registerRequest)
	if err != nil {
		glog.Fatalf("Failed to register handler: %v")
	}
	glog.Infof("Registered iso: %v\n", reply)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		unregisterRequest := &lifecycle.UnregisterRequest{
			Name:  name,
			Token: reply.Token,
		}
		rep, err := client.Unregister(ctx, unregisterRequest)
		if err != nil {
			glog.Fatalf("Failed to unregister handler: %v")
		}
		glog.Infof("Unregistered iso: %v\n", rep)

		os.Exit(1)
	}()
	server := &eventHandler{}
	server.Start(eventHandlerAddress)

}
