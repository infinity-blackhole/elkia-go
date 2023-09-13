package cluster

import (
	"context"

	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
)

type MemoryClusterServerConfig struct {
	Members []*fleetpb.Member
}

func NewMemoryClusterServer(config MemoryClusterServerConfig) *MemoryClusterServer {
	return &MemoryClusterServer{
		members: config.Members,
	}
}

type MemoryClusterServer struct {
	fleetpb.UnimplementedClusterServer
	members []*fleetpb.Member
}

func (s *MemoryClusterServer) MemberAdd(
	ctx context.Context,
	in *fleetpb.MemberAddRequest,
) (*fleetpb.MemberAddResponse, error) {
	s.members = append(s.members, &fleetpb.Member{
		Id:         in.Id,
		WorldId:    in.WorldId,
		ChannelId:  in.ChannelId,
		Name:       in.Name,
		Addresses:  in.Addresses,
		Population: in.Population,
		Capacity:   in.Capacity,
	})
	return &fleetpb.MemberAddResponse{}, nil
}

func (s *MemoryClusterServer) MemberRemove(
	ctx context.Context,
	in *fleetpb.MemberRemoveRequest,
) (*fleetpb.MemberRemoveResponse, error) {
	for i, member := range s.members {
		if member.Id == in.Id {
			s.members = append(s.members[:i], s.members[i+1:]...)
			break
		}
	}
	return &fleetpb.MemberRemoveResponse{}, nil
}

func (s *MemoryClusterServer) MemberUpdate(
	ctx context.Context,
	in *fleetpb.MemberUpdateRequest,
) (*fleetpb.MemberUpdateResponse, error) {
	for _, member := range s.members {
		if member.Id == in.Id {
			if in.WorldId != nil {
				member.WorldId = *in.WorldId
			}
			if in.ChannelId != nil {
				member.ChannelId = *in.ChannelId
			}
			if in.Name != nil {
				member.Name = *in.Name
			}
			if in.Addresses != nil {
				member.Addresses = in.Addresses
			}
			if in.Population != nil {
				member.Population = *in.Population
			}
			if in.Capacity != nil {
				member.Capacity = *in.Capacity
			}
			break
		}
	}
	return &fleetpb.MemberUpdateResponse{}, nil
}

func (s *MemoryClusterServer) MemberList(
	ctx context.Context,
	in *fleetpb.MemberListRequest,
) (*fleetpb.MemberListResponse, error) {
	return &fleetpb.MemberListResponse{
		Members: s.members,
	}, nil
}
