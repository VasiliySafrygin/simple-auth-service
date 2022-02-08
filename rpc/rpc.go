package rpc

import (
	"auth-service/db"
	"auth-service/function"
	pb "auth-service/rpcpb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
}

func (h *Server) GetUserByName(ctx context.Context, request *pb.GetUserByNameRequest) (*pb.GetUserByNameResponse, error) {
	user, err := db.GetUserByName(request.Name)

	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		pbUser := &pb.User{Id: user.Id, Username: user.UserName, Password: user.Password}
		return &pb.GetUserByNameResponse{User: pbUser}, nil
	}

}

func (h *Server) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := db.CreateUser(request.Username, request.FirstName,  request.LastName, request.MiddleName, request.Password)
	if err != nil {
		return nil, err
	} else {
		pbUser := &pb.User{
			Id: user.Id,
			Username: user.UserName,
			FirstName: user.FirstName,
			LastName: user.LastName,
			MiddleName: user.MiddleName,
			Password: user.Password,
		}
		token, _ := db.GetToken(user.Id)
		return &pb.CreateUserResponse{User: pbUser, Token: token}, nil
	}
}

func (h *Server) CheckToken(ctx context.Context, request *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error) {
	userId, err := function.VerifyToken(request.Token)
	//fmt.Println("UserId %s, err %v", userId, err)
	if err != nil {
		return &pb.CheckTokenResponse{Result: false}, err
	} else {
		return &pb.CheckTokenResponse{Result: true, UserId: userId}, nil
	}
}

func (h *Server) mustEmbedUnimplementedAuthServiceServer() {
	//panic("implement me")
}

func StartGRPCServer(address string) {
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not attach listener to port: %v", err)
	}

	//server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &Server{})
	log.Printf("Start gRPC Server %s\n", address)
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("could not start grpc server: %v", err)
		}
	}()
}

//func (h *AuthService) GetUserToken(r *http.Request, args *struct{ Id string }, reply *struct{ Token string }) error {
//	token, err := db.GetToken(args.Id)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		reply.Token = token
//	}
//	return nil
//}
//
