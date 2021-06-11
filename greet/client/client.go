package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/saha/grpc-go-course/greet"
	"github.com/saha/grpc-go-course/greet/greetpb"
	"github.com/saha/grpc-go-course/utils"
)

func doUnary(client greetpb.GreetServiceClient) {
	log.Println("Starting to do a Unary RPC...")
	request := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sumit",
			LastName:  "Saha",
		},
	}
	response, err := client.Greet(context.Background(), request)
	if err != nil {
		log.Fatalf("Error while calling Greet rpc : %v", err)
	}
	log.Printf("Response from greet : %v", response.Result)
}

func doServerStreaming(client greetpb.GreetServiceClient) {
	log.Println("Starting to do a Server Streaming RPC...")
	request := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sumit",
			LastName:  "Saha",
		},
	}
	responseStream, err := client.GreetManyTimes(context.Background(), request)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes rpc : %v", err)
	}

	for {
		response, err := responseStream.Recv()
		if err == io.EOF {
			// End of stream has been reached
			log.Println("Stream finished")
			break
		} else if err != nil {
			log.Fatalf("Error while reading stream : %v", err)
		}
		log.Printf("Response from greet : %v", response.Result)
	}
}

func doClientStreaming(client greetpb.GreetServiceClient) {
	log.Println("Starting to do a Client Streaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sumit",
				LastName:  "Saha",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sumit1",
				LastName:  "Saha1",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sumit2",
				LastName:  "Saha2",
			},
		},
	}
	stream, err := client.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet : %v", err)
	}

	for _, request := range requests {
		log.Printf("Sending request : %v", request)
		if err = stream.Send(request); err != nil {
			log.Printf("Unable to send message : %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet : %v", err)
	}

	log.Printf("LongGreet received response : %v", response)
}

func doBiDirectionalStreaming(client greetpb.GreetServiceClient) {
	log.Println("Starting to do a BiDirectional Streaming RPC...")

	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sumit",
				LastName:  "Saha",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sumit1",
				LastName:  "Saha1",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sumit2",
				LastName:  "Saha2",
			},
		},
	}

	// We create a stream by invoking the client
	stream, err := client.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream in GreetEveryone : %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	//Here we are Sending and receiving messages on the same stream but in different go routine
	go func(requests []*greetpb.GreetEveryoneRequest, stream greetpb.GreetService_GreetEveryoneClient) {
		// Func to send some messages
		for _, req := range requests {
			log.Printf("Request sent from Client  : %v from client \n", req)
			// This send is not a blocking call, on the server there seems a queue if the server is not able to process the data
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("Unable to send request %v", err)
			}
			//time.Sleep(1*time.Second)
		}
		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Unable to close stream %v", err)
		}
	}(requests, stream)

	go func(wg *sync.WaitGroup, stream greetpb.GreetService_GreetEveryoneClient) {
		// Func to receive some messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving stream %v", err)
			}
			log.Printf("Response received at Client : %v \n", res.Result)
		}
		wg.Done()
	}(wg, stream)

	wg.Wait()
}

func doUnaryWithDeadline(client greetpb.GreetServiceClient, timeout time.Duration) {
	log.Println("Starting to do a Unary RPC...")
	request := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sumit",
			LastName:  "Saha",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	response, err := client.GreetWithDeadline(ctx, request)
	if err != nil {

		statusError, ok := status.FromError(err)
		if ok {
			if statusError.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit!, Deadline exceeded")
			} else {
				fmt.Printf("Unexpected error %v \n", statusError)
			}
		} else {
			log.Fatalf("Error while calling Greet rpc : %v", err)
		}
	} else {
		log.Printf("Response from greet : %v", response.Result)
	}

}

func main() {
	utils.SetLogger("logs/greet-client-logs.txt")
	log.Println("Starting gRPC Client")

	var dialOptions []grpc.DialOption
	if greet.UseTLS {
		clientCert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
		if err != nil {
			log.Fatalf("Failed to load client certificate and key. %s.", err)
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
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS13,
			MaxVersion:   tls.VersionTLS13,
		}

		// Create a new TLS credentials based on the TLS configuration
		cred := credentials.NewTLS(tlsConfig)
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(cred))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	clientConnection, err := grpc.Dial(greet.Host, dialOptions...) // With SSL
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer func(clientConnection *grpc.ClientConn) {
		err = clientConnection.Close()
		if err != nil {
		}
	}(clientConnection)

	client := greetpb.NewGreetServiceClient(clientConnection)

	/*doUnary(client)
	fmt.Println()
	doServerStreaming(client)
	fmt.Println()
	doClientStreaming(client)
	fmt.Println()*/
	doBiDirectionalStreaming(client)
	/*fmt.Println()
	doUnaryWithDeadline(client, 5 * time.Second) // should complete
	doUnaryWithDeadline(client, 1 * time.Second) // should timeout*/
}
