package main

import (
	"basic_service/models"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

// Simple web service that just makes a request to another service.
// For testing, it accepts the following arguments to force a response code and
// delay from the other service.
const (
	argForceReturnCode = "force_ret"
	argForceDelay      = "force_delay"
)

type server struct {
	models.UnimplementedBasicServiceServer
}

func (s *server) Request(ctx context.Context, req *models.BasicRequest) (*models.BasicResponse, error) {
	log.Println("received request")
	// make grpc request to 6060
	target := os.Getenv("BASIC_SERVER_GRPC_URL")
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "couldn't connect to client")
	}

	defer conn.Close()
	c := models.NewBasicServiceClient(conn)

	// Contact the server and print out its response.

	return c.Request(ctx, req)
}

func handleRequest3(rw http.ResponseWriter, req *http.Request) {
	log.Println("received request", req.RequestURI)
	// handle forced delay
	delay := req.URL.Query().Get(argForceDelay)
	retCodeValue := req.URL.Query().Get(argForceReturnCode)
	path := req.URL.Path

	// add delay and return code to request to other service
	url := os.Getenv("BASIC_SERVER_URL")
	if path != "" {
		url += path
	}
	if delay != "" {
		url += "?" + argForceDelay + "=" + delay
	}
	if retCodeValue != "" && delay != "" {
		url += "&" + argForceReturnCode + "=" + retCodeValue
	}
	if retCodeValue != "" && delay == "" {
		url += "?" + argForceReturnCode + "=" + retCodeValue
	}

	log.Println("making request to", url)
	// make request to other service
	resp, err := http.Get(url)
	if err != nil {
		log.Println("error making request to other service:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Println("received response from other service")

	// copy response from other service to response writer
	rw.WriteHeader(resp.StatusCode)
	_, err = io.Copy(rw, resp.Body)
	if err != nil {
		return
	}
}

func main() {
	log.Println("Listening on http://localhost:8081")
	go func() {
		http.HandleFunc("/request", handleRequest3)
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":6061")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	models.RegisterBasicServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
