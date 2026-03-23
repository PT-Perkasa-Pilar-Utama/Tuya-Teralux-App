{
    "openapi": "3.1.0",
    "info": {
        "title": "Sensio API",
        "description": "This is the API server for Sensio App. <br> For full documentation, visit <a href=\"/docs\">/docs</a>.",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "servers": [
        {
            "url": "/",
            "description": "Current server (auto-detected)"
        }
    ],
    "paths": {
        "/api/big/device/{mac_address}": {
            "get": {
                "tags": [
                    "08. Common"
                ],
                "summary": "[Common] Fetch device and booking info by MAC address",
                "description": "Fetches specific device information including booking_id, time_start, and time_stop directly from the Big API.",
                "parameters": [
                    {
                        "name": "mac_address",
                        "in": "path",
                        "description": "MAC Address",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "object",
                                                    "additionalProperties": true
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/cache/flush": {
            "delete": {
                "tags": [
                    "08. Common"
                ],
                "summary": "[Common] Flush all cache",
                "description": "Remove all data from the cache storage",
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/devices": {
            "post": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] Create a new device",
                "description": "Register a new device under a terminal",
                "requestBody": {
                    "description": "Device registration data",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.CreateDeviceRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Created",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/devices/statuses": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Device] Get all device statuses",
                "description": "Retrieve a list of all device statuses",
                "parameters": [
                    {
                        "name": "page",
                        "in": "query",
                        "description": "Page number",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "limit",
                        "in": "query",
                        "description": "Items per page",
                        "schema": {
                            "type": "integer"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceStatusListResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/devices/terminal/{terminal_id}": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Device] Get devices by terminal ID",
                "description": "Retrieve all devices associated with a specific terminal",
                "parameters": [
                    {
                        "name": "terminal_id",
                        "in": "path",
                        "description": "Terminal ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "page",
                        "in": "query",
                        "description": "Page number",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "limit",
                        "in": "query",
                        "description": "Items per page",
                        "schema": {
                            "type": "integer"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceListResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/devices/{id}": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Device] Get device by ID",
                "description": "Retrieve device information by device ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "put": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Device] Update a device",
                "description": "Update an existing device's details by ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Updated device data",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.UpdateDeviceRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.UpdateDeviceRequestDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/devices/{id}/status": {
            "put": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Device] Update device status",
                "description": "Update the status value for a specific device by ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Updated status data",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.UpdateDeviceStatusRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceStatusResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.UpdateDeviceStatusRequestDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/devices/{id}/statuses": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Device] Get device statuses by device ID",
                "description": "Retrieve all statuses for a specific device",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "page",
                        "in": "query",
                        "description": "Page number",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "limit",
                        "in": "query",
                        "description": "Items per page",
                        "schema": {
                            "type": "integer"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.DeviceStatusListResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/health": {
            "get": {
                "tags": [
                    "08. Common"
                ],
                "summary": "[Common] Health check endpoint",
                "description": "Check if the application and database are healthy",
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/mail/send": {
            "post": {
                "tags": [
                    "07. Mail"
                ],
                "summary": "[Mail] Send an email using a template",
                "description": "Send an email using a server-side template and specified recipients",
                "requestBody": {
                    "description": "Mail Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.MailSendRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.MailTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/mail/send/mac/{mac_address}": {
            "post": {
                "tags": [
                    "07. Mail"
                ],
                "summary": "[Mail] Send an email by Terminal MAC Address",
                "description": "Looks up customer email by MAC address and sends an email using a template",
                "parameters": [
                    {
                        "name": "mac_address",
                        "in": "path",
                        "description": "Terminal MAC Address",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Mail Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.MailSendRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Email task submitted successfully",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.MailTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/mail/status/{task_id}": {
            "get": {
                "tags": [
                    "07. Mail"
                ],
                "summary": "[Mail] Get email task status",
                "description": "Get the status and result of an email sending task.",
                "parameters": [
                    {
                        "name": "task_id",
                        "in": "path",
                        "description": "Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.MailStatusDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/gemini": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Raw prompt query to Gemini model",
                "description": "Send a raw prompt directly to the Gemini LLM model without RAG orchestration.",
                "requestBody": {
                    "description": "Prompt Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGRawPromptRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGRawPromptResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/groq": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Raw prompt query to Groq model",
                "description": "Send a raw prompt directly to the Groq LLM model without RAG orchestration.",
                "requestBody": {
                    "description": "Prompt Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGRawPromptRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGRawPromptResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/llama/cpp": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Raw prompt query to local Llama.cpp model",
                "description": "Send a raw prompt directly to the local Llama.cpp LLM model without RAG orchestration.",
                "requestBody": {
                    "description": "Prompt Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGRawPromptRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGRawPromptResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/openai": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Raw prompt query to OpenAI model",
                "description": "Send a raw prompt directly to the OpenAI LLM model without RAG orchestration.",
                "requestBody": {
                    "description": "Prompt Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGRawPromptRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGRawPromptResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/orion": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Raw prompt query to Orion model",
                "description": "Send a raw prompt directly to the Orion LLM model without RAG orchestration.",
                "requestBody": {
                    "description": "Prompt Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGRawPromptRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGRawPromptResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/pipeline/job": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [Pipeline] Run unified AI pipeline (Transcribe -> Refine -> Translate -> Summarize)",
                "description": "Submits a single job that orchestrates multiple AI stages.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Source language (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "target_language",
                        "in": "formData",
                        "description": "Target language for translation/summary",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "summarize",
                        "in": "formData",
                        "description": "Whether to generate a summary",
                        "schema": {
                            "type": "boolean"
                        }
                    },
                    {
                        "name": "refine",
                        "in": "formData",
                        "description": "Whether to refine text (grammar/spelling)",
                        "schema": {
                            "type": "boolean"
                        }
                    },
                    {
                        "name": "diarize",
                        "in": "formData",
                        "description": "Whether to diarize speakers",
                        "schema": {
                            "type": "boolean"
                        }
                    },
                    {
                        "name": "context",
                        "in": "formData",
                        "description": "Meeting context (for summary)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "style",
                        "in": "formData",
                        "description": "Summary style",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "date",
                        "in": "formData",
                        "description": "Meeting date",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "location",
                        "in": "formData",
                        "description": "Meeting location",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "participants",
                        "in": "formData",
                        "description": "Comma-separated participants",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "mac_address",
                        "in": "formData",
                        "description": "Device MAC Address",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "Idempotency-Key",
                        "in": "header",
                        "description": "Idempotency key",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.PipelineResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/pipeline/status/{task_id}": {
            "get": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [Pipeline] Get unified pipeline job status",
                "description": "Poll for status and results of a unified pipeline job.",
                "parameters": [
                    {
                        "name": "task_id",
                        "in": "path",
                        "description": "Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.PipelineStatusDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/rag/chat": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [RAG] AI Assistant Chat",
                "description": "Classifies user prompt into Chat or Control and returns appropriate response or redirection. Infrastructure failures (network, timeout, backend processing) return a graceful service-issue response, not an error.",
                "requestBody": {
                    "description": "Chat Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGChatRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGChatResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/rag/control": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [RAG] AI Assistant Control",
                "description": "Processes natural language device control commands.",
                "requestBody": {
                    "description": "Control Request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGControlRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "string"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/rag/summary": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [RAG] Generate meeting minutes summary",
                "description": "Generate meeting minutes summary asynchronously. Returns a Task ID for polling.",
                "parameters": [
                    {
                        "name": "Idempotency-Key",
                        "in": "header",
                        "description": "Idempotency key to deduplicate requests",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Summary request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGSummaryRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "object",
                                                    "additionalProperties": {
                                                        "type": "string"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/rag/translate": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [RAG] Translate text to specified language",
                "description": "Translate text to a target language asynchronously. Returns a Task ID for polling.",
                "parameters": [
                    {
                        "name": "Idempotency-Key",
                        "in": "header",
                        "description": "Idempotency key to deduplicate requests",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Translation request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.RAGRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "object",
                                                    "additionalProperties": {
                                                        "type": "string"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/rag/{task_id}": {
            "get": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [RAG] Get RAG task status",
                "parameters": [
                    {
                        "name": "task_id",
                        "in": "path",
                        "description": "Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.RAGStatusDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/whisper/transcribe": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [Whisper] Transcribe audio file (Unified)",
                "description": "Start transcription of audio file using the configured LLM provider (LLM_PROVIDER). Asynchronous processing. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Language code (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "mac_address",
                        "in": "formData",
                        "description": "Device MAC Address",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "diarize",
                        "in": "formData",
                        "description": "Identify speakers in transcription",
                        "schema": {
                            "type": "boolean"
                        }
                    },
                    {
                        "name": "Idempotency-Key",
                        "in": "header",
                        "description": "Idempotency key to deduplicate requests",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TranscriptionTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "413": {
                        "description": "Request Entity Too Large",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "415": {
                        "description": "Unsupported Media Type",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/models/whisper/transcribe/{transcribe_id}": {
            "get": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [Whisper] Get transcription status (Consolidated)",
                "description": "Get the status and result of any transcription task (Short, Long, or Orion).",
                "parameters": [
                    {
                        "name": "transcribe_id",
                        "in": "path",
                        "description": "Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.AsyncTranscriptionStatusDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/mqtt/credentials/{username}": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] Get MQTT credentials",
                "description": "Get MQTT credentials by username for device authentication",
                "parameters": [
                    {
                        "name": "username",
                        "in": "path",
                        "description": "MQTT Username",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.MQTTCredentialsResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/notification/publish": {
            "post": {
                "tags": [
                    "08. Common"
                ],
                "summary": "[Common] Publish a notification to all terminals in a room",
                "description": "Computes publish_at = datetime_end - interval_time and immediately publishes it to all terminals in the room via MQTT.",
                "requestBody": {
                    "description": "Notification details",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.NotificationPublishRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.NotificationPublishResponse"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/recordings": {
            "get": {
                "tags": [
                    "06. Recordings"
                ],
                "summary": "[Recordings] List all recordings",
                "description": "Get a paginated list of all recordings.",
                "parameters": [
                    {
                        "name": "page",
                        "in": "query",
                        "description": "Page number (default 1)",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "limit",
                        "in": "query",
                        "description": "Items per page (default 10)",
                        "schema": {
                            "type": "integer"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/recordings_dtos.GetAllRecordingsResponseDto"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "post": {
                "tags": [
                    "06. Recordings"
                ],
                "summary": "[Recordings] Save a new recording",
                "description": "Upload an audio file and save its metadata.",
                "parameters": [
                    {
                        "name": "file",
                        "in": "formData",
                        "description": "Audio file",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "mac_address",
                        "in": "formData",
                        "description": "Device MAC Address",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/recordings_dtos.RecordingResponseDto"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/recordings/{id}": {
            "get": {
                "tags": [
                    "06. Recordings"
                ],
                "summary": "[Recordings] Get recording by ID",
                "description": "Retrieve metadata for a specific recording.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Recording ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/recordings_dtos.RecordingResponseDto"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "delete": {
                "tags": [
                    "06. Recordings"
                ],
                "summary": "[Recordings] Delete a recording",
                "description": "Remove a recording and its associated file from the system.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Recording ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/scenes": {
            "get": {
                "tags": [
                    "03. Scenes"
                ],
                "summary": "[Scenes] List all scenes across all Terminal devices",
                "description": "Retrieve all configured scenes grouped under each Terminal device",
                "responses": {
                    "200": {
                        "description": "All scenes grouped by Terminal",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "array",
                                                    "items": {
                                                        "$ref": "#/components/schemas/dtos.TerminalScenesWrapperDTO"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/terminal": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] Get all terminals",
                "description": "Retrieve a list of all registered terminals with optional filtering",
                "parameters": [
                    {
                        "name": "mac_address",
                        "in": "query",
                        "description": "Filter by MAC address",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "name",
                        "in": "query",
                        "description": "Filter by terminal name",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "room_id",
                        "in": "query",
                        "description": "Filter by room ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "page",
                        "in": "query",
                        "description": "Page number",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "limit",
                        "in": "query",
                        "description": "Items per page",
                        "schema": {
                            "type": "integer"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TerminalListResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "post": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] Create a new terminal",
                "description": "Register a new terminal device with MAC address and metadata",
                "requestBody": {
                    "description": "Terminal registration data",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.CreateTerminalRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Created",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.CreateTerminalResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/terminal/mac/{mac}": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Terminal] Get terminal by MAC address",
                "description": "Retrieve terminal information by MAC address",
                "parameters": [
                    {
                        "name": "mac",
                        "in": "path",
                        "description": "Terminal MAC Address",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TerminalSingleResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/terminal/{id}": {
            "get": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Terminal] Get terminal by ID",
                "description": "Retrieve terminal information by terminal ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TerminalSingleResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "put": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Terminal] Update a terminal",
                "description": "Update an existing terminal record's metadata by ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Updated terminal data",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.UpdateTerminalRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TerminalResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.UpdateTerminalRequestDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "delete": {
                "tags": [
                    "02. Terminal"
                ],
                "summary": "[Terminal] [Terminal] Delete a terminal",
                "description": "Soft delete a terminal by ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/terminal/{id}/scenes": {
            "get": {
                "tags": [
                    "03. Scenes"
                ],
                "summary": "[Scenes] [Terminal] List all scenes for a Terminal device",
                "description": "Retrieve a list of all configured scenes for a specific Terminal device, including all actions for each scene.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of scenes with actions",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "array",
                                                    "items": {
                                                        "$ref": "#/components/schemas/dtos.SceneResponseDTO"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "post": {
                "tags": [
                    "03. Scenes"
                ],
                "summary": "[Scenes] [Terminal] Create a new scene",
                "description": "Create a scene with a name and a list of actions (Tuya/MQTT)",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Scene configuration",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.CreateTerminalRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Created",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.SceneIDResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/terminal/{id}/scenes/{scene_id}": {
            "put": {
                "tags": [
                    "03. Scenes"
                ],
                "summary": "[Scenes] [Terminal] Update an existing scene",
                "description": "Update the configuration of a specific scene",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "scene_id",
                        "in": "path",
                        "description": "Scene UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Updated scene configuration",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.UpdateSceneRequestDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.SceneIDResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            },
            "delete": {
                "tags": [
                    "03. Scenes"
                ],
                "summary": "[Scenes] [Terminal] Delete a scene",
                "description": "Remove a specific scene configuration",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "scene_id",
                        "in": "path",
                        "description": "Scene UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Scene deleted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/terminal/{id}/scenes/{scene_id}/control": {
            "get": {
                "tags": [
                    "03. Scenes"
                ],
                "summary": "[Scenes] [Terminal] Apply/Trigger a scene",
                "description": "Trigger all actions defined in a specific scene",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Terminal UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "scene_id",
                        "in": "path",
                        "description": "Scene UUID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Scene applied",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/tuya/auth": {
            "get": {
                "tags": [
                    "01. Tuya"
                ],
                "summary": "[Tuya] Authenticate with Tuya",
                "description": "Authenticates the user and retrieves a Tuya access token",
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TuyaAuthResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ]
            }
        },
        "/api/tuya/devices": {
            "get": {
                "tags": [
                    "01. Tuya"
                ],
                "summary": "[Tuya] Get All Devices",
                "description": "Retrieves a list of all devices in a Merged View (Smart IR remotes merged with Hubs). Sorted alphabetically by Name. For infrared_ac devices, the status array is populated with saved device state (power, temp, mode, wind) or default values if no state exists.",
                "parameters": [
                    {
                        "name": "page",
                        "in": "query",
                        "description": "Page number",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "limit",
                        "in": "query",
                        "description": "Items per page",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "per_page",
                        "in": "query",
                        "description": "Items per page (alias for limit)",
                        "schema": {
                            "type": "integer"
                        }
                    },
                    {
                        "name": "category",
                        "in": "query",
                        "description": "Filter by category",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TuyaDevicesResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/tuya/devices/{id}": {
            "get": {
                "tags": [
                    "01. Tuya"
                ],
                "summary": "[Tuya] [Device] Get Device by ID",
                "description": "Retrieves details of a specific device by its ID",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "remote_id",
                        "in": "query",
                        "description": "Optional Remote sub-device ID",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TuyaDeviceResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/tuya/devices/{id}/commands/ir": {
            "post": {
                "tags": [
                    "01. Tuya"
                ],
                "summary": "[Tuya] [Device] Send IR AC Command",
                "description": "Sends an infrared command (e.g., AC control) to an IR-enabled Tuya device.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Infrared Device ID (Hub/Remote)",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "IR Command Payload",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.TuyaCommandDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "object",
                                                    "additionalProperties": {
                                                        "type": "boolean"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/tuya/devices/{id}/commands/switch": {
            "post": {
                "tags": [
                    "01. Tuya"
                ],
                "summary": "[Tuya] [Device] Send Switch Command",
                "description": "Sends a standard switch command (e.g., toggle power) to a specific Tuya device.",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Command Payload",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/dtos.TuyaCommandDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "type": "object",
                                                    "additionalProperties": {
                                                        "type": "boolean"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/tuya/devices/{id}/sensor": {
            "get": {
                "tags": [
                    "01. Tuya"
                ],
                "summary": "[Tuya] [Device] Get Sensor Data",
                "description": "Retrieves sensor data (temperature, humidity, etc.) for a specific device",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Device ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.SensorDataDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/pipeline/job": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Pipeline] Execute AI pipeline job",
                "description": "Execute a full AI pipeline (transcribe, translate, summarize)",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Source language (default: id)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "target_language",
                        "in": "formData",
                        "description": "Target language (default: en)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "summarize",
                        "in": "formData",
                        "description": "Enable summarization (true/false)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "diarize",
                        "in": "formData",
                        "description": "Enable speaker diarization (true/false)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "refine",
                        "in": "formData",
                        "description": "Enable refinement (true/false)",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "participants",
                        "in": "formData",
                        "description": "Comma-separated participant names",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TaskIDResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/pipeline/status/{task_id}": {
            "get": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Pipeline] Get pipeline status",
                "description": "Get the status of a pipeline job by ID",
                "parameters": [
                    {
                        "name": "task_id",
                        "in": "path",
                        "description": "Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.V1PipelineStatusResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/rag/chat": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [RAG] Chat with model (v1)",
                "description": "Chat using the legacy Python RAG service",
                "requestBody": {
                    "description": "Chat request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/services.RAGRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/rag/control": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [RAG] Device control (v1)",
                "description": "Control devices using the legacy Python RAG service",
                "requestBody": {
                    "description": "Control request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/services.RAGRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/rag/status/{task_id}": {
            "get": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [RAG] Get RAG status (v1)",
                "description": "Get the status of a RAG task by ID",
                "parameters": [
                    {
                        "name": "task_id",
                        "in": "path",
                        "description": "Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/rag/summary": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [RAG] Summarize text (v1)",
                "description": "Summarize text using the legacy Python RAG service",
                "requestBody": {
                    "description": "Summary request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/services.RAGRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/rag/translate": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [RAG] Translate text (v1)",
                "description": "Translate text using the legacy Python RAG service",
                "requestBody": {
                    "description": "Translation request",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/services.RAGRequest"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/whisper/status/{transcribe_id}": {
            "get": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Whisper] Get transcription status (v1)",
                "description": "Get the status of a transcription task by ID",
                "parameters": [
                    {
                        "name": "transcribe_id",
                        "in": "path",
                        "description": "Transcription Task ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/whisper/transcribe": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Whisper] Transcribe audio (v1)",
                "description": "Transcribe an audio file by providing its local path",
                "parameters": [
                    {
                        "name": "audio_path",
                        "in": "formData",
                        "description": "Local path to audio file",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Audio language",
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "diarize",
                        "in": "formData",
                        "description": "Enable speaker diarization",
                        "schema": {
                            "type": "boolean"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/whisper/uploads/sessions": {
            "post": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Whisper] Create upload session (v1)",
                "description": "Initialize a new chunked upload session for a large audio file",
                "requestBody": {
                    "description": "Upload session configuration",
                    "required": true,
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/sensio_domain_models-v1_whisper_dtos.UploadChunkAckDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "201": {
                        "description": "Created",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/sensio_domain_models-v1_whisper_dtos.UploadSessionResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/whisper/uploads/sessions/{id}": {
            "get": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Whisper] Get upload session status (v1)",
                "description": "Get the current status and progress of an upload session",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Session ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/sensio_domain_models-v1_whisper_dtos.UploadSessionResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/v1/models/whisper/uploads/sessions/{id}/chunks/{index}": {
            "put": {
                "tags": [
                    "05. Models-v1"
                ],
                "summary": "[Models-v1] [Whisper] Upload chunk (v1)",
                "description": "Upload a single chunk of an audio file for a session",
                "parameters": [
                    {
                        "name": "id",
                        "in": "path",
                        "description": "Session ID",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "name": "index",
                        "in": "path",
                        "description": "Chunk index",
                        "required": true,
                        "schema": {
                            "type": "integer"
                        }
                    }
                ],
                "requestBody": {
                    "description": "Binary chunk data",
                    "required": true,
                    "content": {
                        "application/octet-stream": {},
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/sensio_domain_models-v1_whisper_dtos.UploadChunkAckDTO"
                            }
                        }
                    }
                },
                "responses": {
                    "200": {
                        "description": "OK",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/sensio_domain_models-v1_whisper_dtos.UploadChunkAckDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/whisper/models/gemini": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Transcribe audio file (Gemini)",
                "description": "Submit audio file for transcription via Gemini. Processing is asynchronous.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Language code (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TranscriptionTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "413": {
                        "description": "Request Entity Too Large",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "415": {
                        "description": "Unsupported Media Type",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/whisper/models/groq": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Transcribe audio file (Groq)",
                "description": "Submit audio file for transcription via Groq. Processing is asynchronous.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Language code (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TranscriptionTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "413": {
                        "description": "Request Entity Too Large",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "415": {
                        "description": "Unsupported Media Type",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/whisper/models/openai": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Transcribe audio file (OpenAI)",
                "description": "Submit audio file for transcription via OpenAI. Processing is asynchronous.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Language code (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TranscriptionTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "413": {
                        "description": "Request Entity Too Large",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "415": {
                        "description": "Unsupported Media Type",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/whisper/models/orion": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] Transcribe audio file (Orion)",
                "description": "Submit audio file for transcription via Orion. Processing is asynchronous.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Language code (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TranscriptionTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "413": {
                        "description": "Request Entity Too Large",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "415": {
                        "description": "Unsupported Media Type",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        },
        "/api/whisper/models/whisper/cpp": {
            "post": {
                "tags": [
                    "04. Models"
                ],
                "summary": "[Models] [Whisper] Transcribe audio file (Whisper.cpp)",
                "description": "Submit audio file for transcription via Whisper.cpp. Processing is asynchronous.",
                "parameters": [
                    {
                        "name": "audio",
                        "in": "formData",
                        "description": "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)",
                        "required": true,
                        "schema": {
                            "type": "file"
                        }
                    },
                    {
                        "name": "language",
                        "in": "formData",
                        "description": "Language code (e.g. id, en)",
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "allOf": [
                                        {
                                            "$ref": "#/components/schemas/dtos.StandardResponse"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data": {
                                                    "$ref": "#/components/schemas/dtos.TranscriptionTaskResponseDTO"
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ValidationErrorResponse"
                                }
                            }
                        }
                    },
                    "413": {
                        "description": "Request Entity Too Large",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "415": {
                        "description": "Unsupported Media Type",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.StandardResponse"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/dtos.ErrorResponse"
                                }
                            }
                        }
                    }
                },
                "security": [
                    {
                        "BearerAuth": []
                    }
                ]
            }
        }
    },
    "components": {
        "schemas": {
            "dtos.ActionDTO": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string"
                    },
                    "device_id": {
                        "type": "string"
                    },
                    "remote_id": {
                        "type": "string"
                    },
                    "topic": {
                        "type": "string"
                    },
                    "value": {}
                }
            },
            "dtos.AsyncTranscriptionResultDTO": {
                "type": "object",
                "properties": {
                    "detected_language": {
                        "type": "string",
                        "example": "id"
                    },
                    "refined_text": {
                        "type": "string",
                        "example": "Hello world"
                    },
                    "transcription": {
                        "type": "string",
                        "example": "Halo dunia"
                    }
                }
            },
            "dtos.AsyncTranscriptionStatusDTO": {
                "type": "object",
                "properties": {
                    "duration_seconds": {
                        "type": "number",
                        "example": 1.5
                    },
                    "error": {
                        "type": "string",
                        "example": "service unavailable"
                    },
                    "expires_at": {
                        "type": "string"
                    },
                    "expires_in_seconds": {
                        "type": "integer"
                    },
                    "result": {
                        "$ref": "#/components/schemas/dtos.AsyncTranscriptionResultDTO"
                    },
                    "started_at": {
                        "type": "string",
                        "example": "2026-02-21T11:00:00Z"
                    },
                    "status": {
                        "type": "string",
                        "example": "completed"
                    },
                    "terminal_id": {
                        "type": "string"
                    },
                    "trigger": {
                        "type": "string",
                        "example": "/api/whisper/models/gemini"
                    }
                }
            },
            "dtos.CreateDeviceRequestDTO": {
                "type": "object",
                "properties": {
                    "id": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    },
                    "terminal_id": {
                        "type": "string"
                    }
                },
                "required": [
                    "id",
                    "name",
                    "terminal_id"
                ]
            },
            "dtos.CreateSceneRequestDTO": {
                "type": "object",
                "properties": {
                    "actions": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.ActionDTO"
                        }
                    },
                    "name": {
                        "type": "string"
                    }
                },
                "required": [
                    "name"
                ]
            },
            "dtos.CreateTerminalRequestDTO": {
                "type": "object",
                "properties": {
                    "device_type_id": {
                        "type": "string"
                    },
                    "mac_address": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    },
                    "room_id": {
                        "type": "string"
                    }
                },
                "required": [
                    "device_type_id",
                    "mac_address",
                    "name",
                    "room_id"
                ]
            },
            "dtos.CreateTerminalResponseDTO": {
                "type": "object",
                "properties": {
                    "mqtt_password": {
                        "type": "string"
                    },
                    "mqtt_username": {
                        "type": "string"
                    },
                    "terminal_id": {
                        "type": "string"
                    }
                }
            },
            "dtos.DeviceListResponseDTO": {
                "type": "object",
                "properties": {
                    "devices": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.DeviceResponseDTO"
                        }
                    },
                    "page": {
                        "type": "integer"
                    },
                    "per_page": {
                        "type": "integer"
                    },
                    "total": {
                        "type": "integer"
                    }
                }
            },
            "dtos.DeviceResponseDTO": {
                "type": "object",
                "properties": {
                    "category": {
                        "type": "string"
                    },
                    "collections": {
                        "type": "string"
                    },
                    "create_time": {
                        "type": "integer"
                    },
                    "created_at": {
                        "type": "string"
                    },
                    "custom_name": {
                        "type": "string"
                    },
                    "gateway_id": {
                        "type": "string"
                    },
                    "icon": {
                        "type": "string"
                    },
                    "id": {
                        "type": "string"
                    },
                    "ip": {
                        "type": "string"
                    },
                    "local_key": {
                        "type": "string"
                    },
                    "model": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    },
                    "product_name": {
                        "type": "string"
                    },
                    "remote_category": {
                        "type": "string"
                    },
                    "remote_id": {
                        "type": "string"
                    },
                    "remote_product_name": {
                        "type": "string"
                    },
                    "terminal_id": {
                        "type": "string"
                    },
                    "update_time": {
                        "type": "integer"
                    },
                    "updated_at": {
                        "type": "string"
                    }
                }
            },
            "dtos.DeviceStatusListResponseDTO": {
                "type": "object",
                "properties": {
                    "device_statuses": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.DeviceStatusResponseDTO"
                        }
                    },
                    "page": {
                        "type": "integer"
                    },
                    "per_page": {
                        "type": "integer"
                    },
                    "total": {
                        "type": "integer"
                    }
                }
            },
            "dtos.DeviceStatusResponseDTO": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string"
                    },
                    "created_at": {
                        "type": "string"
                    },
                    "device_id": {
                        "type": "string"
                    },
                    "updated_at": {
                        "type": "string"
                    },
                    "value": {
                        "type": "string"
                    }
                }
            },
            "dtos.ErrorResponse": {
                "type": "object",
                "properties": {
                    "data": {},
                    "message": {
                        "type": "string",
                        "example": "An error occurred"
                    },
                    "status": {
                        "type": "boolean",
                        "example": false
                    }
                }
            },
            "dtos.MQTTCredentialsResponseDTO": {
                "type": "object",
                "properties": {
                    "password": {
                        "type": "string",
                        "example": "secret_password"
                    },
                    "username": {
                        "type": "string",
                        "example": "device_001"
                    }
                }
            },
            "dtos.MailSendRequestDTO": {
                "type": "object",
                "properties": {
                    "attachment_path": {
                        "type": "string",
                        "example": "/uploads/reports/019c981c-d7ec-7dd2-a642-9f6e5dbe7e37.pdf"
                    },
                    "data": {
                        "type": "object",
                        "additionalProperties": true
                    },
                    "subject": {
                        "type": "string",
                        "example": "Notification"
                    },
                    "template": {
                        "type": "string",
                        "example": "test"
                    },
                    "to": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "example": [
                            "user@example.com"
                        ]
                    }
                },
                "required": [
                    "subject",
                    "to"
                ]
            },
            "dtos.MailStatusDTO": {
                "type": "object",
                "properties": {
                    "duration_seconds": {
                        "type": "number"
                    },
                    "error": {
                        "type": "string",
                        "example": "smtp auth failed"
                    },
                    "expires_at": {
                        "type": "string"
                    },
                    "expires_in_seconds": {
                        "type": "integer"
                    },
                    "result": {
                        "type": "string",
                        "example": "Email sent to user@example.com"
                    },
                    "started_at": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string",
                        "example": "completed"
                    },
                    "trigger": {
                        "type": "string"
                    }
                }
            },
            "dtos.MailTaskResponseDTO": {
                "type": "object",
                "properties": {
                    "task_id": {
                        "type": "string",
                        "example": "mail-abc-123"
                    },
                    "task_status": {
                        "type": "string",
                        "example": "pending"
                    }
                }
            },
            "dtos.NotificationPublishRequest": {
                "type": "object",
                "properties": {
                    "datetime_end": {
                        "type": "string",
                        "example": "2026-03-17T14:00:00+07:00"
                    },
                    "interval_time": {
                        "type": "integer",
                        "example": 15
                    },
                    "room_id": {
                        "type": "string",
                        "example": "123"
                    }
                },
                "required": [
                    "datetime_end",
                    "room_id"
                ]
            },
            "dtos.NotificationPublishResponse": {
                "type": "object",
                "properties": {
                    "publish_at": {
                        "type": "string",
                        "example": "2026-03-17T13:45:00+07:00"
                    },
                    "published_count": {
                        "type": "integer",
                        "example": 2
                    },
                    "published_topics": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "example": [
                            "[\"users/AA:BB:CC:DD:EE:FF/notification\"]"
                        ]
                    },
                    "room_id": {
                        "type": "string",
                        "example": "123"
                    }
                }
            },
            "dtos.PipelineResponseDTO": {
                "type": "object",
                "properties": {
                    "task_id": {
                        "type": "string"
                    },
                    "task_status": {
                        "$ref": "#/components/schemas/dtos.PipelineStatusDTO"
                    }
                }
            },
            "dtos.PipelineStageStatus": {
                "type": "object",
                "properties": {
                    "duration_seconds": {
                        "type": "number"
                    },
                    "error": {
                        "type": "string"
                    },
                    "result": {},
                    "started_at": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string",
                        "description": "pending, processing, completed, failed, skipped",
                        "example": "pending"
                    }
                }
            },
            "dtos.PipelineStatusDTO": {
                "type": "object",
                "properties": {
                    "duration_seconds": {
                        "type": "number"
                    },
                    "expires_at": {
                        "type": "string"
                    },
                    "expires_in_seconds": {
                        "type": "integer"
                    },
                    "overall_status": {
                        "type": "string",
                        "description": "pending, processing, completed, failed",
                        "example": "processing"
                    },
                    "stages": {
                        "type": "object",
                        "additionalProperties": {
                            "$ref": "#/components/schemas/dtos.PipelineStageStatus"
                        }
                    },
                    "started_at": {
                        "type": "string"
                    },
                    "task_id": {
                        "type": "string"
                    }
                }
            },
            "dtos.RAGChatRequestDTO": {
                "type": "object",
                "properties": {
                    "language": {
                        "type": "string",
                        "example": "id"
                    },
                    "prompt": {
                        "type": "string",
                        "example": "Nyalakan AC"
                    },
                    "request_id": {
                        "type": "string"
                    },
                    "terminal_id": {
                        "type": "string",
                        "example": "tx-1"
                    },
                    "uid": {
                        "type": "string",
                        "description": "Must be Tuya UID (never MAC/terminal identity)",
                        "example": "sg1765..."
                    }
                },
                "required": [
                    "prompt",
                    "terminal_id"
                ]
            },
            "dtos.RAGChatResponseDTO": {
                "type": "object",
                "properties": {
                    "instance_id": {
                        "type": "string",
                        "description": "Server start time"
                    },
                    "is_blocked": {
                        "type": "boolean"
                    },
                    "is_control": {
                        "type": "boolean"
                    },
                    "redirect": {
                        "$ref": "#/components/schemas/dtos.RedirectDTO"
                    },
                    "request_id": {
                        "type": "string",
                        "description": "Tracking ID"
                    },
                    "response": {
                        "type": "string"
                    },
                    "source": {
                        "type": "string",
                        "description": "e.g., \"MQTT\", \"HTTP\""
                    }
                }
            },
            "dtos.RAGControlRequestDTO": {
                "type": "object",
                "properties": {
                    "prompt": {
                        "type": "string",
                        "example": "Nyalakan AC"
                    },
                    "terminal_id": {
                        "type": "string",
                        "example": "tx-1"
                    }
                },
                "required": [
                    "prompt",
                    "terminal_id"
                ]
            },
            "dtos.RAGRawPromptRequestDTO": {
                "type": "object",
                "properties": {
                    "prompt": {
                        "type": "string",
                        "example": "Hello, how are you?"
                    }
                },
                "required": [
                    "prompt"
                ]
            },
            "dtos.RAGRawPromptResponseDTO": {
                "type": "object",
                "properties": {
                    "duration_seconds": {
                        "type": "number"
                    },
                    "error": {
                        "type": "string"
                    },
                    "result": {
                        "type": "string"
                    },
                    "started_at": {
                        "type": "string"
                    },
                    "status": {
                        "type": "string",
                        "example": "completed"
                    },
                    "trigger": {
                        "type": "string"
                    }
                }
            },
            "dtos.RAGRequestDTO": {
                "type": "object",
                "properties": {
                    "language": {
                        "type": "string",
                        "example": "id"
                    },
                    "mac_address": {
                        "type": "string",
                        "example": "AA:BB:CC:DD:EE:FF"
                    },
                    "text": {
                        "type": "string",
                        "example": "Ini adalah transkrip panjang dari rapat teknis..."
                    }
                },
                "required": [
                    "text"
                ]
            },
            "dtos.RAGStatusDTO": {
                "type": "object",
                "properties": {
                    "agenda_context": {
                        "type": "string"
                    },
                    "body": {},
                    "duration_seconds": {
                        "type": "number",
                        "example": 2.5
                    },
                    "endpoint": {
                        "type": "string"
                    },
                    "error": {
                        "type": "string",
                        "example": "gemini api returned status 503"
                    },
                    "execution_result": {
                        "description": "holds the response from the fetched endpoint"
                    },
                    "expires_at": {
                        "type": "string"
                    },
                    "expires_in_seconds": {
                        "type": "integer"
                    },
                    "headers": {
                        "type": "object",
                        "additionalProperties": {
                            "type": "string"
                        }
                    },
                    "language": {
                        "type": "string"
                    },
                    "mac_address": {
                        "type": "string"
                    },
                    "meeting_context": {
                        "type": "string"
                    },
                    "method": {
                        "type": "string"
                    },
                    "pdf_url": {
                        "type": "string"
                    },
                    "result": {
                        "type": "string",
                        "example": "The meeting discussed..."
                    },
                    "started_at": {
                        "type": "string",
                        "example": "2026-02-21T11:00:00Z"
                    },
                    "status": {
                        "type": "string",
                        "example": "completed"
                    },
                    "summary": {
                        "type": "string",
                        "description": "Alias for Result in summary tasks"
                    },
                    "trigger": {
                        "type": "string",
                        "example": "/api/rag/summary"
                    }
                }
            },
            "dtos.RAGSummaryRequestDTO": {
                "type": "object",
                "properties": {
                    "context": {
                        "type": "string",
                        "example": "technical meeting"
                    },
                    "date": {
                        "type": "string"
                    },
                    "language": {
                        "type": "string",
                        "description": "\"id\" or \"en\"",
                        "example": "id"
                    },
                    "location": {
                        "type": "string",
                        "example": "Meeting Room A"
                    },
                    "mac_address": {
                        "type": "string",
                        "example": "AA:BB:CC:DD:EE:FF"
                    },
                    "participants": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "example": [
                            "[\"Alice\"",
                            " \"Bob\"]"
                        ]
                    },
                    "style": {
                        "type": "string",
                        "description": "e.g., \"minutes\", \"executive\"",
                        "example": "minutes"
                    },
                    "text": {
                        "type": "string",
                        "example": "This is a long transcript of a technical meeting..."
                    }
                },
                "required": [
                    "text"
                ]
            },
            "dtos.RedirectDTO": {
                "type": "object",
                "properties": {
                    "body": {},
                    "endpoint": {
                        "type": "string"
                    },
                    "method": {
                        "type": "string"
                    }
                }
            },
            "dtos.SceneIDResponseDTO": {
                "type": "object",
                "properties": {
                    "scene_id": {
                        "type": "string"
                    }
                }
            },
            "dtos.SceneItemDTO": {
                "type": "object",
                "properties": {
                    "actions": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.ActionDTO"
                        }
                    },
                    "id": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    }
                }
            },
            "dtos.SceneResponseDTO": {
                "type": "object",
                "properties": {
                    "actions": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.ActionDTO"
                        }
                    },
                    "id": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    },
                    "terminal_id": {
                        "type": "string"
                    }
                }
            },
            "dtos.SendMailByMacRequestDTO": {
                "type": "object",
                "properties": {
                    "attachment_path": {
                        "type": "string",
                        "example": "/uploads/reports/019c981c-d7ec-7dd2-a642-9f6e5dbe7e37.pdf"
                    },
                    "data": {
                        "type": "object",
                        "additionalProperties": true
                    },
                    "subject": {
                        "type": "string",
                        "example": "Booking Confirmation"
                    },
                    "template": {
                        "type": "string",
                        "example": "test"
                    }
                },
                "required": [
                    "subject"
                ]
            },
            "dtos.SensorDataDTO": {
                "type": "object",
                "properties": {
                    "battery_percentage": {
                        "type": "integer"
                    },
                    "humidity": {
                        "type": "integer"
                    },
                    "status_text": {
                        "type": "string"
                    },
                    "temp_unit": {
                        "type": "string"
                    },
                    "temperature": {
                        "type": "number"
                    }
                }
            },
            "dtos.StandardResponse": {
                "type": "object",
                "properties": {
                    "data": {},
                    "details": {},
                    "message": {
                        "type": "string",
                        "example": "Success"
                    },
                    "status": {
                        "type": "boolean",
                        "example": true
                    }
                }
            },
            "dtos.TaskIDResponseDTO": {
                "type": "object",
                "properties": {
                    "task_id": {
                        "type": "string",
                        "example": "task-123"
                    }
                }
            },
            "dtos.TerminalListResponseDTO": {
                "type": "object",
                "properties": {
                    "page": {
                        "type": "integer"
                    },
                    "per_page": {
                        "type": "integer"
                    },
                    "terminal": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.TerminalResponseDTO"
                        }
                    },
                    "total": {
                        "type": "integer"
                    }
                }
            },
            "dtos.TerminalResponseDTO": {
                "type": "object",
                "properties": {
                    "ai_provider": {
                        "type": "string"
                    },
                    "created_at": {
                        "type": "string"
                    },
                    "device_type_id": {
                        "type": "string"
                    },
                    "id": {
                        "type": "string"
                    },
                    "mac_address": {
                        "type": "string"
                    },
                    "mqtt_password": {
                        "type": "string"
                    },
                    "mqtt_username": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    },
                    "room_id": {
                        "type": "string"
                    },
                    "updated_at": {
                        "type": "string"
                    }
                }
            },
            "dtos.TerminalScenesDTO": {
                "type": "object",
                "properties": {
                    "scenes": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.SceneItemDTO"
                        }
                    },
                    "terminal_id": {
                        "type": "string"
                    }
                }
            },
            "dtos.TerminalScenesWrapperDTO": {
                "type": "object",
                "properties": {
                    "terminal": {
                        "$ref": "#/components/schemas/dtos.TerminalScenesDTO"
                    }
                }
            },
            "dtos.TerminalSingleResponseDTO": {
                "type": "object",
                "properties": {
                    "terminal": {
                        "$ref": "#/components/schemas/dtos.TerminalResponseDTO"
                    }
                }
            },
            "dtos.TranscriptionTaskResponseDTO": {
                "type": "object",
                "properties": {
                    "recording_id": {
                        "type": "string",
                        "example": "uuid-v4"
                    },
                    "task_id": {
                        "type": "string",
                        "example": "abc-123"
                    },
                    "task_status": {
                        "type": "string",
                        "example": "pending"
                    }
                }
            },
            "dtos.TuyaAuthResponseDTO": {
                "type": "object",
                "properties": {
                    "access_token": {
                        "type": "string"
                    },
                    "expire_time": {
                        "type": "integer"
                    },
                    "refresh_token": {
                        "type": "string"
                    },
                    "uid": {
                        "type": "string"
                    }
                }
            },
            "dtos.TuyaCommandDTO": {
                "type": "object"
            },
            "dtos.TuyaDeviceDTO": {
                "type": "object",
                "properties": {
                    "category": {
                        "type": "string"
                    },
                    "collections": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.TuyaDeviceDTO"
                        }
                    },
                    "create_time": {
                        "type": "integer"
                    },
                    "custom_name": {
                        "type": "string"
                    },
                    "gateway_id": {
                        "type": "string"
                    },
                    "icon": {
                        "type": "string"
                    },
                    "id": {
                        "type": "string"
                    },
                    "ip": {
                        "type": "string"
                    },
                    "local_key": {
                        "type": "string"
                    },
                    "model": {
                        "type": "string"
                    },
                    "name": {
                        "type": "string"
                    },
                    "online": {
                        "type": "boolean"
                    },
                    "product_name": {
                        "type": "string"
                    },
                    "remote_category": {
                        "type": "string"
                    },
                    "remote_id": {
                        "type": "string"
                    },
                    "remote_product_name": {
                        "type": "string"
                    },
                    "status": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.TuyaDeviceStatusDTO"
                        }
                    },
                    "update_time": {
                        "type": "integer"
                    }
                }
            },
            "dtos.TuyaDeviceResponseDTO": {
                "type": "object",
                "properties": {
                    "device": {
                        "$ref": "#/components/schemas/dtos.TuyaDeviceDTO"
                    }
                }
            },
            "dtos.TuyaDeviceStatusDTO": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string"
                    },
                    "value": {}
                }
            },
            "dtos.TuyaDevicesResponseDTO": {
                "type": "object",
                "properties": {
                    "current_page_count": {
                        "type": "integer"
                    },
                    "devices": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.TuyaDeviceDTO"
                        }
                    },
                    "page": {
                        "type": "integer"
                    },
                    "per_page": {
                        "type": "integer"
                    },
                    "total": {
                        "type": "integer"
                    },
                    "total_devices": {
                        "type": "integer"
                    }
                }
            },
            "dtos.TuyaIRACCommandDTO": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string",
                        "example": "temp"
                    },
                    "remote_id": {
                        "type": "string",
                        "example": "ir-remote-001"
                    },
                    "value": {
                        "type": "integer",
                        "example": 24
                    }
                },
                "required": [
                    "code",
                    "remote_id"
                ]
            },
            "dtos.UpdateDeviceRequestDTO": {
                "type": "object",
                "properties": {
                    "name": {
                        "type": "string",
                        "example": "Updated Device Name"
                    }
                }
            },
            "dtos.UpdateDeviceStatusRequestDTO": {
                "type": "object"
            },
            "dtos.UpdateSceneRequestDTO": {
                "type": "object",
                "properties": {
                    "actions": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.ActionDTO"
                        }
                    },
                    "name": {
                        "type": "string",
                        "example": "Evening Mode"
                    }
                },
                "required": [
                    "name"
                ]
            },
            "dtos.UpdateTerminalRequestDTO": {
                "type": "object",
                "properties": {
                    "ai_provider": {
                        "type": "string",
                        "example": "openai"
                    },
                    "device_type_id": {
                        "type": "string",
                        "example": "hub-type-002"
                    },
                    "mac_address": {
                        "type": "string",
                        "example": "AA:BB:CC:DD:EE:FF"
                    },
                    "name": {
                        "type": "string",
                        "example": "Updated Hub Name"
                    },
                    "room_id": {
                        "type": "string",
                        "example": "room-456"
                    }
                }
            },
            "dtos.V1PipelineStatusResponseDTO": {
                "type": "object",
                "properties": {
                    "error": {
                        "type": "string"
                    },
                    "refined_text": {
                        "type": "string",
                        "example": "Hello, world."
                    },
                    "status": {
                        "type": "string",
                        "example": "completed"
                    },
                    "summary": {
                        "type": "string",
                        "example": "A greeting to the world."
                    },
                    "task_id": {
                        "type": "string",
                        "example": "pipeline_task_123"
                    },
                    "transcript": {
                        "type": "string",
                        "example": "Hello world"
                    },
                    "translated": {
                        "type": "string",
                        "example": "Halo dunia"
                    }
                }
            },
            "dtos.ValidationErrorDetailDTO": {
                "type": "object",
                "properties": {
                    "field": {
                        "type": "string",
                        "example": "username"
                    },
                    "message": {
                        "type": "string",
                        "example": "is required"
                    }
                }
            },
            "dtos.ValidationErrorResponse": {
                "type": "object",
                "properties": {
                    "data": {},
                    "details": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/dtos.ValidationErrorDetailDTO"
                        }
                    },
                    "message": {
                        "type": "string",
                        "example": "Validation Error"
                    },
                    "status": {
                        "type": "boolean",
                        "example": false
                    }
                }
            },
            "recordings_dtos.GetAllRecordingsResponseDto": {
                "type": "object",
                "properties": {
                    "limit": {
                        "type": "integer"
                    },
                    "page": {
                        "type": "integer"
                    },
                    "recordings": {
                        "type": "array",
                        "items": {
                            "$ref": "#/components/schemas/recordings_dtos.RecordingResponseDto"
                        }
                    },
                    "total": {
                        "type": "integer"
                    }
                }
            },
            "recordings_dtos.RecordingResponseDto": {
                "type": "object",
                "properties": {
                    "audio_url": {
                        "type": "string"
                    },
                    "created_at": {
                        "type": "string"
                    },
                    "filename": {
                        "type": "string"
                    },
                    "id": {
                        "type": "string"
                    },
                    "original_name": {
                        "type": "string"
                    }
                }
            },
            "sensio_domain_models-v1_whisper_dtos.CreateUploadSessionRequest": {
                "type": "object",
                "properties": {
                    "chunk_size_bytes": {
                        "type": "integer"
                    },
                    "file_name": {
                        "type": "string"
                    },
                    "mime_type": {
                        "type": "string"
                    },
                    "total_size_bytes": {
                        "type": "integer"
                    }
                },
                "required": [
                    "file_name",
                    "total_size_bytes"
                ]
            },
            "sensio_domain_models-v1_whisper_dtos.UploadChunkAckDTO": {
                "type": "object",
                "properties": {
                    "is_duplicate": {
                        "type": "boolean"
                    },
                    "received_bytes": {
                        "type": "integer"
                    },
                    "received_chunks": {
                        "type": "integer"
                    },
                    "state": {
                        "type": "string"
                    }
                }
            },
            "sensio_domain_models-v1_whisper_dtos.UploadSessionResponseDTO": {
                "type": "object",
                "properties": {
                    "chunk_size_bytes": {
                        "type": "integer"
                    },
                    "expires_at": {
                        "type": "integer"
                    },
                    "missing_ranges": {
                        "type": "array",
                        "description": "e.g. [\"0-2\", \"5\"]",
                        "items": {
                            "type": "string"
                        }
                    },
                    "received_bytes": {
                        "type": "integer"
                    },
                    "session_id": {
                        "type": "string"
                    },
                    "state": {
                        "type": "string",
                        "description": "uploading, ready, consumed, aborted, expired"
                    },
                    "total_chunks": {
                        "type": "integer"
                    },
                    "total_size_bytes": {
                        "type": "integer"
                    }
                }
            },
            "services.RAGRequest": {
                "type": "object",
                "properties": {
                    "language": {
                        "type": "string"
                    },
                    "mac_address": {
                        "type": "string"
                    },
                    "text": {
                        "type": "string"
                    }
                }
            }
        },
        "securitySchemes": {
            "ApiKeyAuth": {
                "type": "apiKey",
                "name": "X-API-KEY",
                "in": "header",
                "scheme": "apiKey"
            },
            "BearerAuth": {
                "type": "apiKey",
                "description": "Type \"Bearer\" followed by a space and JWT token.",
                "name": "Authorization",
                "in": "header",
                "scheme": "apiKey"
            }
        }
    },
    "tags": [
        {
            "name": "01. Tuya",
            "description": "Tuya authentication and device control endpoints"
        },
        {
            "name": "02. Terminal",
            "description": "Terminal and device management endpoints"
        },
        {
            "name": "03. Scenes",
            "description": "Scene management and control endpoints"
        },
        {
            "name": "04. Models",
            "description": "AI Model access endpoints (Speech, RAG, Pipeline) - Unified domain"
        },
        {
            "name": "05. Models-v1",
            "description": "New AI Model endpoints (Whisper, RAG, Pipeline) - v1 API"
        },
        {
            "name": "06. Recordings",
            "description": "Recordings management endpoints"
        },
        {
            "name": "07. Mail",
            "description": "Mail service endpoints"
        },
        {
            "name": "08. Common",
            "description": "Common endpoints (Health, Cache, External APIs)"
        }
    ]
}