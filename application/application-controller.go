package application

import (
	"context"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
)

type SaaSApplication struct {
	ID            string
	Name          string
	Description   string
	Visiblity     string
	Driver        string
	Configuration map[string]string
	URL           string

	InternalURL string
	Location    string
}

// THIS SHOULD NOT BE IN CLOUDY?
type Team struct {
	ID          string
	Name        string
	Description string
	Visiblity   string
	ParentID    string
	GroupID     string

	ApplicationConfigs map[string]interface{}
}

var SaasAplicationControllers = cloudy.NewProviderRegistry[SaasAplicationController]()

type SaasAplicationController interface {
	// ApplyTeamConfig applys a configuration to the application. This encompassed add and removes, etc
	ApplyTeamConfig(ctx context.Context, app *SaaSApplication, cfg interface{}) error

	// Adds a member to the team
	AddMember(ctx context.Context, app *SaaSApplication, uid string) error

	// Removes a member from the team
	RemoveMember(ctx context.Context, app *SaaSApplication, uid string) error

	// Set Members
	SetMembers(ctx context.Context, app *SaaSApplication, members []*models.User) error

	// Archives / Deactivates team
	DeactivateTeam(ctx context.Context, app *SaaSApplication) error

	// Permenantly deletes team
	DeleteTeam(ctx context.Context, app *SaaSApplication) error
}

type SaasApplicationTeamController interface {
	// ApplyTeamConfig applys a configuration to the application. This encompassed add and removes, etc
	ApplyTeamConfig(ctx context.Context, app *SaaSApplication, cfg interface{}) error

	// Adds a member to the team
	AddMember(ctx context.Context, app *SaaSApplication, uid string) error

	// Removes a member from the team
	RemoveMember(ctx context.Context, app *SaaSApplication, uid string) error

	// Set Members
	SetMembers(ctx context.Context, app *SaaSApplication, members []*models.User) error

	// Archives / Deactivates team
	DeactivateTeam(ctx context.Context, app *SaaSApplication) error

	// Permenantly deletes team
	DeleteTeam(ctx context.Context, app *SaaSApplication) error
}
