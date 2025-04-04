// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/projects": {
            "get": {
                "description": "Retrieves a paginated list of projects with their database connections",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "projects"
                ],
                "summary": "List all projects",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Page number (default: 1)",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Number of items per page (default: 10, max: 100)",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/schemas.PageResponse-schemas_ProjectResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new project with name and description",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "projects"
                ],
                "summary": "Create a project",
                "parameters": [
                    {
                        "description": "Project JSON",
                        "name": "project",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schemas.ProjectRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/schemas.ProjectResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "schemas.DBConnectionRequest": {
            "type": "object",
            "required": [
                "dbname",
                "host",
                "password",
                "port",
                "user"
            ],
            "properties": {
                "dbname": {
                    "type": "string"
                },
                "host": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "port": {
                    "type": "integer"
                },
                "user": {
                    "type": "string"
                }
            }
        },
        "schemas.DBConnectionResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "dbname": {
                    "type": "string"
                },
                "host": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "port": {
                    "type": "integer"
                },
                "updated_at": {
                    "type": "string"
                },
                "user": {
                    "type": "string"
                }
            }
        },
        "schemas.PageResponse-schemas_ProjectResponse": {
            "type": "object",
            "properties": {
                "entries": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/schemas.ProjectResponse"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "page": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "schemas.ProjectRequest": {
            "type": "object",
            "required": [
                "description",
                "name",
                "source",
                "target"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "source": {
                    "$ref": "#/definitions/schemas.DBConnectionRequest"
                },
                "target": {
                    "$ref": "#/definitions/schemas.DBConnectionRequest"
                }
            }
        },
        "schemas.ProjectResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "source": {
                    "$ref": "#/definitions/schemas.DBConnectionResponse"
                },
                "target": {
                    "$ref": "#/definitions/schemas.DBConnectionResponse"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Mammoth Server API",
	Description:      "This is a database (PostgreSQL) schema comparsion and migration.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
