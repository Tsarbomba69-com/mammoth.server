package controllers

import (
	"net/http"
	"strconv"

	"github.com/Tsarbomba69-com/mammoth.server/mappers"
	"github.com/Tsarbomba69-com/mammoth.server/models"
	"github.com/Tsarbomba69-com/mammoth.server/repositories"
	"github.com/Tsarbomba69-com/mammoth.server/schemas"
	"github.com/Tsarbomba69-com/mammoth.server/services"
	"github.com/Tsarbomba69-com/mammoth.server/utils"
	"github.com/gin-gonic/gin"
)

// CreateProject creates a new project
// @Summary Create a project
// @Description Create a new project with name and description
// @Tags projects
// @Accept json
// @Produce json
// @Param project body schemas.ProjectRequest true "Project JSON"
// @Success 201 {object} schemas.ProjectResponse
// @Failure 400 {object} map[string]any
// @Router /api/v1/projects [post]
func CreateProject(c *gin.Context) {
	var input schemas.ProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := mappers.ProjectToModel(input)
	repositories.Context.Create(&project)
	c.JSON(http.StatusCreated, mappers.ProjectToResponse(project))
}

// GetProjects retrieves a paginated list of projects
// @Summary List all projects
// @Description Retrieves a paginated list of projects with their database connections
// @Tags projects
// @Accept  json
// @Produce  json
// @Param   page   query     int false "Page number (default: 1)"
// @Param   limit  query     int false "Number of items per page (default: 10, max: 100)"
// @Success 200    {object}  schemas.PageResponse[schemas.ProjectResponse]
// @Failure 400    {object}  map[string]any
// @Router  /api/v1/projects [get]
func GetProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var projects []models.Project
	var total int64
	offset := (page - 1) * limit
	repositories.Context.Preload("Source").Preload("Target").Limit(limit).Offset(offset).Find(&projects)
	repositories.Context.Model(&models.Project{}).Count(&total)
	var projectResponses []schemas.ProjectResponse = []schemas.ProjectResponse{}

	for _, project := range projects {
		projectResponses = append(projectResponses, mappers.ProjectToResponse(project))
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"page":    page,
		"limit":   limit,
		"entries": projectResponses,
	})
}

// MigrateProject initiates database migration for a specific project
// @Summary Migrate database
// @Description Initiates a migration from the source database to the target database for the specified project
// @Tags projects
// @Accept  json
// @Produce  json
// @Param   id   path      string  true  "Project ID"
// @Success 200  {object}  map[string]any
// @Failure 400  {object}  map[string]any
// @Failure 404  {object}  map[string]any
// @Failure 500  {object}  map[string]any
// @Router  /api/v1/projects/{id}/migrate [post]
func MigrateProject(c *gin.Context) {
	projectID := c.Param("id")
	var project models.Project
	var sourceSchema []services.TableSchema
	var targetSchema []services.TableSchema
	var sw = utils.NewStopwatch()

	if err := repositories.Context.Preload("Source").Preload("Target").First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	sw.Start("Connect Project")
	var source, target, err = project.ConnectForProject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to databases"})
		return
	}

	sw.Stop("Connect Project")
	sw.Start("Dump Source Schema")
	sourceSchema, err = services.DumpSchemaAST(source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dump schema"})
		return
	}

	sw.Stop("Dump Source Schema")
	sw.Start("Dump Target Schema")
	targetSchema, err = services.DumpSchemaAST(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dump schema"})
		return
	}

	sw.Stop("Dump Target Schema")
	sw.Start("Compare Schemas")
	diff := services.CompareSchemas(sourceSchema, targetSchema)
	sw.Stop("Compare Schemas")
	sw.Start("Generate Migration Script")
	script := services.Generate(diff)
	sw.Stop("Generate Migration Script")
	sw.PrintTable()
	c.JSON(http.StatusOK, gin.H{"up": script.Up, "down": script.Down})
}
