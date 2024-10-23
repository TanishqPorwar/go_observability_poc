package main

import (
	"basic_service/models"
	"context"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Simple web service that just returns Ok to any path.
// For testing, it accepts the following arguments in order to change the
// response:
const (
	argForceReturnCode = "force_ret"
	argForceDelay      = "force_delay"
)

var (
	brokerUrl = os.Getenv("KAFKA_BROKER_URL")
	topic     = os.Getenv("KAFKA_TOPIC")
	producer  = &kafkago.Writer{
		Addr:         kafkago.TCP(brokerUrl),
		BatchSize:    1,
		BatchTimeout: 1,
		Async:        true,
	}
)

type server struct {
	models.UnimplementedBasicServiceServer
}

func (s *server) Request(ctx context.Context, req *models.BasicRequest) (*models.BasicResponse, error) {

	msg := fmt.Sprintf("Received grpc request with delay: %d, return code: %d", req.ForceDelay, req.ForceRet)
	// handle forced delay
	if req.ForceDelay > 0 {
		time.Sleep(time.Duration(req.ForceDelay) * time.Millisecond)
	}

	// handle forced response code
	retCode := codes.OK
	if req.ForceRet > 0 {
		retCode = codes.Code(req.ForceRet)
		return nil, status.Error(retCode, "forced error")
	}

	err := producer.WriteMessages(ctx, kafkago.Message{
		Value: []byte(msg),
		Topic: topic,
	})
	if err != nil {
		return nil, err
	}

	return &models.BasicResponse{}, nil

}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	log.Println("received request", req.RequestURI)
	ctx := req.Context()
	delay := req.URL.Query().Get(argForceDelay)
	resCode := req.URL.Query().Get(argForceReturnCode)

	msg := fmt.Sprintf("Received http request with delay: %s, return code: %s", delay, resCode)

	// handle forced delay
	if d, err := strconv.Atoi(delay); err == nil {
		time.Sleep(time.Duration(d) * time.Millisecond)
	}

	// handle forced response code
	retCode := http.StatusOK
	if r, err := strconv.Atoi(resCode); err == nil {
		retCode = r
	}

	err := producer.WriteMessages(ctx, kafkago.Message{
		Value: []byte(msg),
		Topic: topic,
	})
	if err != nil {
		log.Println("error writing to kafka:", err)
		retCode = http.StatusInternalServerError
	}

	rw.WriteHeader(retCode)
}

func main() {
	log.Println("Listening on http://localhost:8080")
	go func() {
		http.HandleFunc("/request", handleRequest)
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":6060")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	models.RegisterBasicServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
