package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"log-service/data"
	"log-service/logs"
	"net"
)

type LogServer struct {
	logs.UnimplementedLoggerServer
	data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, request *logs.LogRequest) (*logs.LogResponse, error) {
	input := request.GetLogEntry()
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Message: "Failed"}
		return res, err
	}

	res := &logs.LogResponse{Message: "Success logged!"}
	return res, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", ":"+gRpcPort)
	if err != nil {
		log.Println("Failed to listen:", err)
	}

	server := grpc.NewServer()
	logs.RegisterLoggerServer(server, &LogServer{Models: app.Models})

	log.Println("Starting gRPC server on port", gRpcPort)

	err = server.Serve(lis)
	if err != nil {
		log.Println("2 Failed to listen:", err)
	}
}
