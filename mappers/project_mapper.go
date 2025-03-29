package mappers

import (
	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/schemas"
)

func DBConnectionToModel(conn schemas.DBConnectionRequest) models.DBConnection {
	return models.DBConnection{
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
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
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		Source:      DBConnectionToResponse(&project.Source),
		Target:      DBConnectionToResponse(&project.Target),
	}
}

func DBConnectionToResponse(model *models.DBConnection) schemas.DBConnectionResponse {
	return schemas.DBConnectionResponse{
		ID:     model.ID,
		Host:   model.Host,
		Port:   model.Port,
		User:   model.User,
		DBName: model.DBName,
	}
}
