package main

import (
	pb "cloud-worker/proto"
	"context"
	"encoding/json"
	"fmt"
	"net"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedWorkerServer
}

var host = "8080"

func receive(event cloudevents.Event) {
	// do something with event.
	fmt.Printf("%s", event)
}

func main() {
	client, err := cloudevents.NewClientHTTP()
	log.Fatal(err)
	event := cloudevents.NewEvent()
	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	if result := client.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}

	client.StartReceiver(context.Background(), receive)
	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// init log entry
	logrusEntry := log.WithField("log", "cloud-worker")
	// Shared options for the logger, with a custom gRPC code to log level function.
	customFunc := func(code codes.Code) log.Level {
		if code == codes.OK {
			return log.InfoLevel
		}
		return log.DebugLevel
	}
	opts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(customFunc),
	}
	grpcSvr := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_logrus.UnaryServerInterceptor(logrusEntry, opts...),
			unaryServerInterceptor,
		),
	)

	pb.RegisterWorkerServer(grpcSvr, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcSvr.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("started!")
}

func InitEvent() {

}

// 注册拦截器给log实例中增加request-id的field,并打印请求包和返回包
func unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ctxlogrus.AddFields(ctx, log.Fields{})
	resp, err = handler(ctx, req)
	reqJson, _ := json.Marshal(req)
	respJson, _ := json.Marshal(resp)
	ctxlogrus.Extract(ctx).Infof("request: < %v > \n response: < %v >", string(reqJson), string(respJson))
	return resp, err
}
