// Code generated by swaggo/swag. DO NOT EDIT.

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
        "/check-username": {
            "get": {
                "description": "아이디 중복확인 시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "회원가입 /user"
                ],
                "summary": "중복확인",
                "parameters": [
                    {
                        "type": "string",
                        "description": "중복체크 할 아이디",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 1,이미 있는 아이디 -1",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/find-password": {
            "post": {
                "description": "비밀번호 찾기 시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "비밀번호 찾기 /user"
                ],
                "summary": "비밀번호 찾기",
                "parameters": [
                    {
                        "description": "정보 dto",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/core.FindPasswordRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 JWT 토큰 반환/ -1: 번호 인증안됨 / -2: 일치하는 회원 없음",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/find-username": {
            "post": {
                "description": "아이디찾기 시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "아이디 찾기 /user"
                ],
                "summary": "아이디 찾기",
                "parameters": [
                    {
                        "description": "정보 dto",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/core.FindUsernameRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 아이디 반환/ -1: 번호 인증안됨 / -2: 일치하는 회원 없음",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/get-user": {
            "post": {
                "description": "내 정보 조회시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "회원조회 /user"
                ],
                "summary": "유저 조회",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer {jwt_token}",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 유저 객체 반환",
                        "schema": {
                            "$ref": "#/definitions/core.UserResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "아이디/비밀번호 로그인 시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "로그인 /user"
                ],
                "summary": "일반로그인",
                "parameters": [
                    {
                        "description": "요청 DTO",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/core.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 JWT 토큰 반환/-1 :아이디 또는 비밀번호 불일치",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/send-code/{number}": {
            "post": {
                "description": "휴대전화 인증번호 발송시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "인증번호 /user"
                ],
                "summary": "인증번호 발송",
                "parameters": [
                    {
                        "type": "string",
                        "description": "휴대번호",
                        "name": "number",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 1 반환",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/set-user": {
            "post": {
                "description": "회원 정보 변경 시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "회원수정 /user"
                ],
                "summary": "회원 데이터 변경",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer {jwt_token}",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "요청 DTO - 업데이트 할 데이터",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/core.SetUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 200 반환",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/sign-in": {
            "post": {
                "description": "회원가입 정보 입력 완료 후 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "회원가입 /user"
                ],
                "summary": "회원가입",
                "parameters": [
                    {
                        "description": "요청 DTO",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/core.SignInRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 1, 휴대폰 인증 안함 -1",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환 ",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/verify-code": {
            "post": {
                "description": "인증번호 인증시 호출",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "인증번호 /user"
                ],
                "summary": "인증번호 인증",
                "parameters": [
                    {
                        "description": "요청 DTO",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/core.VerifyRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "성공시 1 반환 코드불일치 -1",
                        "schema": {
                            "$ref": "#/definitions/core.BasicResponse"
                        }
                    },
                    "400": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "요청 처리 실패시 오류 메시지 반환",
                        "schema": {
                            "$ref": "#/definitions/core.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "core.BasicResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                }
            }
        },
        "core.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "core.FindPasswordRequest": {
            "type": "object",
            "properties": {
                "phone": {
                    "type": "string",
                    "example": "01000000000"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "core.FindUsernameRequest": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "phone": {
                    "type": "string",
                    "example": "01000000000"
                }
            }
        },
        "core.LoginRequest": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "core.SetUserRequest": {
            "type": "object",
            "properties": {
                "addr": {
                    "type": "string"
                },
                "addr_detail": {
                    "type": "string"
                },
                "disable_type": {
                    "type": "integer"
                },
                "is_agree": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "visit_purpose": {
                    "type": "integer"
                }
            }
        },
        "core.SignInRequest": {
            "type": "object",
            "properties": {
                "addr": {
                    "type": "string"
                },
                "addr_detail": {
                    "type": "string"
                },
                "birth": {
                    "type": "string",
                    "example": "yyyy-mm-dd"
                },
                "disable_type": {
                    "type": "integer"
                },
                "gender": {
                    "type": "boolean"
                },
                "is_agree": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "sns_id": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                },
                "visit_purpose": {
                    "type": "integer"
                }
            }
        },
        "core.UserResponse": {
            "type": "object",
            "properties": {
                "addr": {
                    "type": "string"
                },
                "addr_detail": {
                    "type": "string"
                },
                "birth": {
                    "type": "string"
                },
                "disable_type": {
                    "type": "integer"
                },
                "gender": {
                    "type": "boolean"
                },
                "is_agree": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                },
                "visit_purpose": {
                    "type": "integer"
                }
            }
        },
        "core.VerifyRequest": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string",
                    "example": "인증번호 6자리"
                },
                "phone_number": {
                    "type": "string",
                    "example": "01000000000"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
