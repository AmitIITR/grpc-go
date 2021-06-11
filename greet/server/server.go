package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/saha/grpc-go-course/greet"
	"github.com/saha/grpc-go-course/greet/greetpb"
	"github.com/saha/grpc-go-course/utils"
)

type server struct {
	greetpb.UnimplementedGreetServiceServer
}

func (*server) Greet(ctx context.Context, request *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("Greet function invoked with %v \n", request)
	greetRequest := request.GetGreeting()
	if greetRequest == nil {
		return nil, nil
	}
	firstName := greetRequest.GetFirstName()
	lastName := greetRequest.GetLastName()
	greetResponse := &greetpb.GreetResponse{
		Result: "Hello " + firstName + " " + lastName,
	}
	return greetResponse, nil
}

func (*server) GreetManyTimes(request *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Printf("GreetManyTimes function invoked with %v \n", request)
	greetRequest := request.GetGreeting()
	if greetRequest == nil {
		return nil
	}
	firstName := greetRequest.GetFirstName()
	lastName := greetRequest.GetLastName()

	for i := 0; i < 10; i++ {
		response := &greetpb.GreetManyTimesResponse{
			Result: "Hello " + firstName + " " + lastName + " " + strconv.Itoa(i),
		}
		_ = stream.Send(response)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	log.Printf("LongGreet function invoked with streaming request \n")
	result := ""
	for {
		request, err := stream.Recv()
		log.Printf("Received request : %v", request)
		if err == io.EOF {
			response := &greetpb.LongGreetResponse{
				Result: result,
			}
			log.Printf("Sending response : %v", result)
			return stream.SendAndClose(response)
		} else if err != nil {
			log.Fatalf("Error while reading client Stream %v", err)
		}
		firstName := request.GetGreeting().GetFirstName()
		lastName := request.GetGreeting().GetLastName()
		result += "Hello " + firstName + " " + lastName + "! "
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Printf("GreetEveryone function invoked with streaming request \n")

	for {
		request, recvErr := stream.Recv()
		if recvErr == io.EOF {
			return nil
		}
		if recvErr != nil {
			log.Fatalf("Error while reading client Stream %v", recvErr)
		}
		log.Printf("Request received at server : %v \n", request)

		response := processRequestForBidirectionalStreams(request)
		sendErr := stream.Send(response)
		if sendErr != nil {
			log.Fatalf("Error while sending resposne to client %v", sendErr)
		}
		log.Printf("Response sent from server : %v \n\n", response)
	}
}

func processRequestForBidirectionalStreams(request *greetpb.GreetEveryoneRequest) *greetpb.GreetEveryoneResponse {
	log.Printf("Request being processed at server : %v \n", request)
	time.Sleep(10 * time.Second)
	firstName := request.GetGreeting().GetFirstName()
	lastName := request.GetGreeting().GetLastName()
	result := "Hello " + firstName + " " + lastName + "! "
	response := &greetpb.GreetEveryoneResponse{
		Result: result,
	}
	return response
}

func (*server) GreetWithDeadline(ctx context.Context, request *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("Greet function invoked with %v \n", request)
	for i := 0; i < 3 ; i++ {
		if ctx.Err() == context.Canceled {
			log.Println("The client cancelled the request")
			return nil, status.Error(codes.Canceled,"The client cancelled the request")
		}
		time.Sleep(1 * time.Second)
	}
	greetRequest := request.GetGreeting()
	if greetRequest == nil {
		return nil, nil
	}
	firstName := greetRequest.GetFirstName()
	lastName := greetRequest.GetLastName()
	greetResponse := &greetpb.GreetResponse{
		Result: "Hello " + firstName + " " + lastName,
	}
	return greetResponse, nil
}

func main() {
	utils.SetLogger("logs/greet-server-logs.txt")

	log.Println("Starting Greet gRPC Server")

	lis, err := net.Listen(greet.Protocol, greet.Host)

	if err != nil {
		log.Fatal("Failed to listen :- ", err)
	}

	var serverOptions []grpc.ServerOption
	if greet.UseTLS {
		// Load the server certificate and its key
		serverCert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
		if err != nil {
			log.Fatalf("Failed to load server certificate and key. %s.", err)
		}

		// Load the CA certificate
		trustedCert, err := ioutil.ReadFile("certs/cacert.pem")
		if err != nil {
			log.Fatalf("Failed to load trusted certificate. %s.", err)
		}

		// Put the CA certificate to certificate pool
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(trustedCert) {
			log.Fatalf("Failed to append trusted certificate to certificate pool. %s.", err)
		}

		// Create the TLS configuration
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			RootCAs:      certPool,
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   tls.VersionTLS13,
		}

		// Create a new TLS credentials based on the TLS configuration
		cred := credentials.NewTLS(tlsConfig)
		serverOptions = append(serverOptions, grpc.Creds(cred))
	}

	grpcServer := grpc.NewServer(serverOptions...)
	localServer := server{}

	greetpb.RegisterGreetServiceServer(grpcServer, &localServer)
	reflection.Register(grpcServer)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to server :- ", err)
	}
}
