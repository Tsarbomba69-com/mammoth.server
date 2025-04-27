package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Tsarbomba69-com/mammoth.server/controllers"
	"github.com/Tsarbomba69-com/mammoth.server/repositories"
	"github.com/Tsarbomba69-com/mammoth.server/schemas"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	dialector := postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)
	return gormDB, mock
}

func TestCreateProject(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	originalDB := repositories.Context
	defer func() { repositories.Context = originalDB }()
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatal("Error loading .env file")
	}

	t.Run("Success - Project created successfully", func(t *testing.T) {
		// Arrange
		gormDB, mock := setupTestDB(t)
		originalDB := repositories.Context
		repositories.Context = gormDB
		defer func() { repositories.Context = originalDB }()
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "db_connections"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`INSERT INTO "db_connections"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectQuery(`INSERT INTO "projects"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
		mock.ExpectCommit()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := schemas.ProjectRequest{
			Name:        "Test Project",
			Description: "A test project",
			Source: schemas.DBConnectionRequest{
				Host:     "Test Source",
				DBName:   "database",
				Port:     2355,
				User:     "user",
				Password: "password",
			},
			Target: schemas.DBConnectionRequest{
				Host:     "Test Target",
				DBName:   "database",
				Port:     2351,
				User:     "user",
				Password: "password",
			},
		}
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		req := httptest.NewRequest("POST", "/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Act
		controllers.CreateProject(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		var response schemas.ProjectResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, requestBody.Name, response.Name)
		assert.Equal(t, requestBody.Description, response.Description)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Error - Invalid JSON input", func(t *testing.T) {
		// Arrange - No DB mocking needed for this test
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/projects", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		// Act
		controllers.CreateProject(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
