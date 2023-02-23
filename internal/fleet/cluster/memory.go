package cluster

import (
	"context"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
)

type MemoryClusterServerConfig struct {
	Members []*fleet.Member
}

func NewMemoryClusterServer(config MemoryClusterServerConfig) *MemoryClusterServer {
	return &MemoryClusterServer{
		members: config.Members,
	}
}

type MemoryClusterServer struct {
	fleet.UnimplementedClusterServer
	members []*fleet.Member
}

func (s *MemoryClusterServer) MemberAdd(
	ctx context.Context,
	in *fleet.MemberAddRequest,
) (*fleet.MemberAddResponse, error) {
	s.members = append(s.members, &fleet.Member{
		Id:         in.Id,
		WorldId:    in.WorldId,
		ChannelId:  in.ChannelId,
		Name:       in.Name,
		Address:    in.Address,
		Population: in.Population,
		Capacity:   in.Capacity,
	})
	return &fleet.MemberAddResponse{}, nil
}

func (s *MemoryClusterServer) MemberRemove(
	ctx context.Context,
	in *fleet.MemberRemoveRequest,
) (*fleet.MemberRemoveResponse, error) {
	for i, member := range s.members {
		if member.Id == in.Id {
			s.members = append(s.members[:i], s.members[i+1:]...)
			break
		}
	}
	return &fleet.MemberRemoveResponse{}, nil
}

func (s *MemoryClusterServer) MemberUpdate(
	ctx context.Context,
	in *fleet.MemberUpdateRequest,
) (*fleet.MemberUpdateResponse, error) {
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
			if in.Address != nil {
				member.Address = *in.Address
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
	return &fleet.MemberUpdateResponse{}, nil
}

func (s *MemoryClusterServer) MemberList(
	ctx context.Context,
	in *fleet.MemberListRequest,
) (*fleet.MemberListResponse, error) {
	return &fleet.MemberListResponse{
		Members: s.members,
	}, nil
}
