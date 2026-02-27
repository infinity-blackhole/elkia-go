package lobby

import (
	"context"

	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"gorm.io/gorm"
)

type Character struct {
	gorm.Model
	Id             string
	Class          eventing.CharacterClass
	HairColor      eventing.CharacterHairColor
	HairStyle      eventing.CharacterHairStyle
	Faction        eventing.Faction
	Reputation     int32
	Dignity        int32
	Compliment     int32
	Health         int32
	Mana           int32
	Name           string
	HeroExperience int32
	HeroLevel      int32
	JobExperience  int32
	JobLevel       int32
	Experience     int32
	Level          int32
}

type LobbyServerConfig struct {
	DB *gorm.DB
}

func NewLobbyServer(config LobbyServerConfig) *LobbyServer {
	return &LobbyServer{
		db: config.DB,
	}
}

type LobbyServer struct {
	eventing.UnimplementedLobbyServer
	db *gorm.DB
}

func (s *LobbyServer) CharacterList(
	ctx context.Context,
	in *eventing.CharacterListRequest,
) (*eventing.CharacterListResponse, error) {
	var dbChar []Character
	if err := s.db.
		Where("id = ?", in.IdentityId).
		Find(&dbChar).Error; err != nil {
		return nil, err
	}
	var characters []*eventing.Character
	for _, character := range characters {
		characters = append(characters, &eventing.Character{
			Id:             character.Id,
			Class:          character.Class,
			HairColor:      character.HairColor,
			HairStyle:      character.HairStyle,
			Faction:        character.Faction,
			Reputation:     character.Reputation,
			Dignity:        character.Dignity,
			Compliment:     character.Compliment,
			Health:         character.Health,
			Mana:           character.Mana,
			Name:           character.Name,
			HeroExperience: character.HeroExperience,
			HeroLevel:      character.HeroLevel,
			JobExperience:  character.JobExperience,
			JobLevel:       character.JobLevel,
			Experience:     character.Experience,
			Level:          character.Level,
		})
	}
	return &eventing.CharacterListResponse{
		CharacterEvents: characters,
	}, nil
}
