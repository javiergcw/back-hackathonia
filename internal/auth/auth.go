package auth

import (
	"encoding/json"
	"os"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

func LoadUsers(path string) ([]domain.User, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var users []domain.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func IdentifyUser(users []domain.User, phone, profileId string) *domain.User {
	for i := range users {
		if phone != "" && users[i].Phone == phone {
			return &users[i]
		}
		if profileId != "" && users[i].ProfileID == profileId {
			return &users[i]
		}
	}
	return nil
}

func GetAllowedTags(user *domain.User) []string {
	if user == nil {
		return []string{"publico", "general"}
	}
	return user.AllowedTags
}

func CanViewChunk(user *domain.User, chunkTags []string) bool {
	if user == nil {
		return containsTag(chunkTags, "publico") || containsTag(chunkTags, "general")
	}

	for _, allowed := range user.AllowedTags {
		if allowed == "*" {
			return true
		}
	}

	for _, chunkTag := range chunkTags {
		for _, allowed := range user.AllowedTags {
			if chunkTag == allowed {
				return true
			}
		}
	}

	return false
}

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

func DefaultUser() *domain.User {
	return &domain.User{
		ID:          "guest",
		Nombre:      "Invitado",
		Phone:       "",
		Role:        domain.RolePublico,
		ProfileID:   "",
		AllowedTags: []string{"publico", "general"},
	}
}