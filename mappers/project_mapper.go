package mappers

import (
	"os"

	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/schemas"
	"github.com/Tsarbomba69-com/mammoth.server/utils"
)

func DBConnectionToModel(conn schemas.DBConnectionRequest) models.DBConnection {
	pass, err := utils.Encrypt([]byte(os.Getenv("ENCRYPTION_KEY")), conn.Password)
	if err != nil {
		panic(err) // Handle error appropriately in production code
	}
	return models.DBConnection{
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: pass,
		DBName:   conn.DBName,
	}
}

func ProjectToModel(request schemas.ProjectRequest) models.Project {
	return models.Project{
		Name:        request.Name,
		Description: request.Description,
		Source:      DBConnectionToModel(request.Source),
		Target:      DBConnectionToModel(request.Target),
	}
}

// Convert Project Model to ProjectResponse
func ProjectToResponse(project models.Project) schemas.ProjectResponse {
	return schemas.ProjectResponse{
		ID:          project.ID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		Name:        project.Name,
		Description: project.Description,
		Source:      DBConnectionToResponse(&project.Source),
		Target:      DBConnectionToResponse(&project.Target),
	}
}

func DBConnectionToResponse(model *models.DBConnection) schemas.DBConnectionResponse {
	return schemas.DBConnectionResponse{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Host:      model.Host,
		Port:      model.Port,
		User:      model.User,
		DBName:    model.DBName,
	}
}
