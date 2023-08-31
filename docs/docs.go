// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Peter Androsov",
            "url": "http://t.me/nervous_void",
            "email": "androsov.p.v@gmail.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/create_segment": {
            "post": {
                "description": "creates new segment",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Segments"
                ],
                "summary": "creates new segment",
                "parameters": [
                    {
                        "description": "fraction — optional",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/segment.RequestSegmentSlug"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "created",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "bad input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "something went wrong",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/delete_segment": {
            "delete": {
                "description": "deletes existing segment",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Segments"
                ],
                "summary": "deletes existing segment",
                "parameters": [
                    {
                        "description": "The input struct",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/segment.RequestSegmentSlug"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "deleted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "bad input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "something went wrong",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/get_user_history": {
            "get": {
                "description": "receive report on user segments assignments and unassignments within the given dates",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "History"
                ],
                "summary": "receive report on user segments assignments and unassignments",
                "parameters": [
                    {
                        "description": "The input struct",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/history.Request"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/history.ReportResponse"
                        }
                    },
                    "400": {
                        "description": "bad input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "something went wrong",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/get_user_segments": {
            "get": {
                "description": "receive segments assigned to user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Segments"
                ],
                "summary": "receive segments assigned to user",
                "parameters": [
                    {
                        "description": "The input struct",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/segment.RequestUserID"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/segment.UserSegments"
                        }
                    },
                    "400": {
                        "description": "bad input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "something went wrong",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/update_user_segments": {
            "post": {
                "description": "assign and unassign segments from user",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Segments"
                ],
                "summary": "assign and unassign segments from user",
                "parameters": [
                    {
                        "description": "The input struct",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/segment.RequestUpdateSegments"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "assigned and unassigned",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "bad input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "something went wrong",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "history.ReportResponse": {
            "type": "object",
            "properties": {
                "csv_url": {
                    "type": "string"
                }
            }
        },
        "history.Request": {
            "type": "object",
            "properties": {
                "end_date": {
                    "type": "string"
                },
                "start_date": {
                    "type": "string"
                },
                "user_id": {
                    "type": "integer"
                }
            }
        },
        "segment.RequestSegmentSlug": {
            "type": "object",
            "properties": {
                "fraction": {
                    "type": "integer"
                },
                "segment_slug": {
                    "type": "string"
                }
            }
        },
        "segment.RequestUpdateSegments": {
            "type": "object",
            "properties": {
                "assign_segments": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "unassign_segments": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "user_id": {
                    "type": "integer"
                }
            }
        },
        "segment.RequestUserID": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "integer"
                }
            }
        },
        "segment.UserSegments": {
            "type": "object",
            "properties": {
                "segments": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "user_id": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Dynamic User Segmentation Service API",
	Description:      "Avito Tech backend trainee assignment 2023",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
