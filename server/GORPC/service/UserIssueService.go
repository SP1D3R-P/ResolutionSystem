package issues

import (
	"context"

	pb "example.com/ResulationSystem/server/GORPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserIssueSeviceServer struct {
	pb.UnimplementedUserIssueServiceServer
}

func (s *UserIssueSeviceServer) PostIssue(context.Context, *pb.Issue) (*pb.PostIssueResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method PostIssue not implemented")
}
func (s *UserIssueSeviceServer) GetIssuesByFeatureName(*pb.Feature, grpc.ServerStreamingServer[pb.Issue]) error {
	return status.Error(codes.Unimplemented, "method GetIssuesByFeatureName not implemented")
}
func (s *UserIssueSeviceServer) GetIssueByTitle(context.Context, *pb.IssueTitle) (*pb.Issue, error) {
	return nil, status.Error(codes.Unimplemented, "method GetIssueByTitle not implemented")
}
func (s *UserIssueSeviceServer) GetIssueById(context.Context, *pb.IssueId) (*pb.Issue, error) {
	return nil, status.Error(codes.Unimplemented, "method GetIssueById not implemented")
}
