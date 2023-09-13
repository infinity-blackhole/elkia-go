package lobby

import (
	"context"

	world "go.shikanime.studio/elkia/pkg/api/world/v1alpha1"
	"gorm.io/gorm"
)

type Character struct {
	gorm.Model
	Id             string
	Class          world.CharacterClass
	HairColor      world.CharacterHairColor
	HairStyle      world.CharacterHairStyle
	Faction        world.Faction
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
	world.UnimplementedLobbyServer
	db *gorm.DB
}

func (s *LobbyServer) ListCharacter(
	ctx context.Context,
	in *world.ListCharacterCommand,
) (*world.ListCharacterEvent, error) {
	var dbChar []Character
	if err := s.db.
		Where("id = ?", in.IdentityId).
		Find(&dbChar).Error; err != nil {
		return nil, err
	}
	var characters []*world.Character
	for _, character := range characters {
		characters = append(characters, &world.Character{
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
	return &world.ListCharacterEvent{
		Characters: characters,
	}, nil
}
