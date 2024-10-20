package main

import (
	"basic_service/models"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
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

type server struct {
	models.UnimplementedBasicServiceServer
}

func (s *server) Request(ctx context.Context, req *models.BasicRequest) (*models.BasicResponse, error) {

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

	return &models.BasicResponse{}, nil

}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	log.Println("received request", req.RequestURI)
	delay := req.URL.Query().Get(argForceDelay)

	// handle forced delay
	if d, err := strconv.Atoi(delay); err == nil {
		time.Sleep(time.Duration(d) * time.Millisecond)
	}

	// handle forced response code
	resCode := req.URL.Query().Get(argForceReturnCode)
	retCode := http.StatusOK
	if r, err := strconv.Atoi(resCode); err == nil {
		retCode = r
	}

	rw.WriteHeader(retCode)
}

func main() {
	log.Println("Listening on http://localhost:8080")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(handleRequest)))
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
