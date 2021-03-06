 /*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心检索引擎
 * (BlueKing-IAM-Search-Engine) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

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
        "/api/v1/batch-search": {
            "post": {
                "security": [
                    {
                        "AppCode": []
                    },
                    {
                        "AppSecret": []
                    }
                ],
                "description": "batch search the subjects who have the permission of that system/action/resource",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "api"
                ],
                "summary": "batch search subjects by system/action/resource",
                "operationId": "api-batch-search",
                "parameters": [
                    {
                        "description": "the list request",
                        "name": "params",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/types.SearchRequest"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/full-sync": {
            "post": {
                "security": [
                    {
                        "AppCode": []
                    },
                    {
                        "AppSecret": []
                    }
                ],
                "description": "trigger iam search engine full sync task",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "api"
                ],
                "summary": "trigger iam search engine full sync task",
                "operationId": "api-full-sync",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/search": {
            "post": {
                "security": [
                    {
                        "AppCode": []
                    },
                    {
                        "AppSecret": []
                    }
                ],
                "description": "search the subjects who have the permission of that system/action/resource",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "api"
                ],
                "summary": "search subjects by system/action/resource",
                "operationId": "api-search",
                "parameters": [
                    {
                        "description": "the list request",
                        "name": "params",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.SearchRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/stats": {
            "get": {
                "security": [
                    {
                        "AppCode": []
                    },
                    {
                        "AppSecret": []
                    }
                ],
                "description": "get iam search engine stats",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "api"
                ],
                "summary": "get iam search engine stats",
                "operationId": "api-stats",
                "parameters": [
                    {
                        "type": "string",
                        "description": "System ID",
                        "name": "system",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Action ID",
                        "name": "action",
                        "in": "path"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    }
                }
            }
        },
        "/healthz": {
            "get": {
                "description": "/healthz to make sure the server is health",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "basic"
                ],
                "summary": "healthz for server health check",
                "operationId": "healthz",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "/ping to get response from iam, make sure the server is alive",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "basic"
                ],
                "summary": "ping-pong for alive test",
                "operationId": "ping",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    }
                }
            }
        },
        "/version": {
            "get": {
                "description": "/version to get the version of iam",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "basic"
                ],
                "summary": "version for identify",
                "operationId": "version",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        },
                        "headers": {
                            "X-Request-Id": {
                                "type": "string",
                                "description": "the request id"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "gin.H": {
            "type": "object",
            "additionalProperties": true
        },
        "types.Action": {
            "type": "object",
            "required": [
                "id"
            ],
            "properties": {
                "id": {
                    "type": "string",
                    "example": "edit"
                }
            }
        },
        "types.Resource": {
            "type": "array",
            "items": {
                "$ref": "#/definitions/types.ResourceNode"
            }
        },
        "types.ResourceNode": {
            "type": "object",
            "required": [
                "attribute",
                "id",
                "system",
                "type"
            ],
            "properties": {
                "attribute": {
                    "type": "object",
                    "additionalProperties": true
                },
                "id": {
                    "type": "string",
                    "example": "framework"
                },
                "system": {
                    "type": "string",
                    "example": "bk_paas"
                },
                "type": {
                    "type": "string",
                    "example": "app"
                }
            }
        },
        "types.SearchRequest": {
            "type": "object",
            "required": [
                "action",
                "resource",
                "subject_type",
                "system"
            ],
            "properties": {
                "action": {
                    "type": "object",
                    "$ref": "#/definitions/types.Action"
                },
                "limit": {
                    "description": "! we don't support pagination, we can only fetch limit subjects at once",
                    "type": "integer",
                    "example": 10
                },
                "nowTimestamp": {
                    "type": "integer"
                },
                "resource": {
                    "type": "object",
                    "$ref": "#/definitions/types.Resource"
                },
                "subject_type": {
                    "type": "string",
                    "example": "all"
                },
                "system": {
                    "type": "string",
                    "example": "bk_paas"
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
	Version:     "1.0",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "IAM-Search-Engine API",
	Description: "蓝鲸权限中心后台 engine 服务 API 文档",
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
