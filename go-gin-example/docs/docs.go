// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2022-05-02 14:05:32.187663 +0800 CST m=+0.037349251

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v2/articles": {
            "get": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "获取文章列表",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "TagID",
                        "name": "tag_id",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "新建文章",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "TagID",
                        "name": "tag_id",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Title",
                        "name": "title",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Desc",
                        "name": "desc",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Content",
                        "name": "content",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "CreatedBy",
                        "name": "created_by",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "CoverImageUrl",
                        "name": "cover_image_url",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/api/v2/articles/{id}": {
            "get": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "获取指定文章",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "更新指定文章",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "TagID",
                        "name": "tag_id",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Title",
                        "name": "title",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Desc",
                        "name": "desc",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Content",
                        "name": "content",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "ModifiedBy",
                        "name": "modified_by",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "CoverImageUrl",
                        "name": "cover_image_url",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "删除指定文章",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/api/v2/tags": {
            "get": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "获取标签列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name",
                        "name": "name",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "新建文章标签",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "CreatedBy",
                        "name": "created_by",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/api/v2/tags/export": {
            "post": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "导出标签（Excel）",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name",
                        "name": "name",
                        "in": "formData"
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/api/v2/tags/import": {
            "post": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "导入标签（Excel）",
                "parameters": [
                    {
                        "type": "file",
                        "description": "Tags Excel File",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/api/v2/tags/{id}": {
            "put": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "更新指定标签",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Name",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "ModifiedBy",
                        "name": "modified_by",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "State",
                        "name": "state",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "x-token": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "删除指定标签",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization",
                        "name": "x-token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/auth": {
            "post": {
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "获取 token",
                "parameters": [
                    {
                        "type": "string",
                        "description": "username",
                        "name": "username",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "password",
                        "name": "password",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        },
        "/upload": {
            "post": {
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "上传图片",
                "parameters": [
                    {
                        "type": "file",
                        "description": "Image File",
                        "name": "image",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "app.Response": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {
                    "type": "object"
                },
                "msg": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
