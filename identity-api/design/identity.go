package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("identity", func() {
	Title("Identity Service")
	Description("User registration, authentication and token validation")

	Server("identity", func() {
		Host("localhost", func() {
			URI("http://localhost:8081")
			URI("grpc://localhost:9081")
		})
		Services("identity")
	})
})

var User = ResultType("application/vnd.identity.user", func() {
	TypeName("User")
	Attributes(func() {
		Field(1, "id", String, "User identifier")
		Field(2, "email", String, "Email address")
		Field(3, "display_name", String, "Display name")
		Field(4, "created_at", String, "Creation timestamp", func() {
			Format(FormatDateTime)
		})
		Required("id", "email", "display_name", "created_at")
	})
	View("default", func() {
		Attribute("id")
		Attribute("email")
		Attribute("display_name")
		Attribute("created_at")
	})
})

var TokenResult = Type("TokenResult", func() {
	Field(1, "access_token", String, "JWT access token")
	Field(2, "expires_in", Int, "Token expiry window in seconds")
	Required("access_token", "expires_in")
})

var UnauthorizedError = Type("UnauthorizedError", func() {
	Field(1, "message", String, "description of the failure")
	Field(2, "id", String, "error identifier", func() {
		Example("identity:unauthorized")
	})
	Field(3, "temporary", Boolean, "true if the error is temporary")
	Field(4, "timeout", Boolean, "true if the error is retryable")
	Required("message")
})

var NotFoundError = Type("NotFoundError", func() {
	Field(1, "message", String, "description of the failure")
	Field(2, "id", String, "error identifier", func() {
		Example("identity:not_found")
	})
	Field(3, "temporary", Boolean)
	Field(4, "timeout", Boolean)
	Required("message")
})

var ValidationResult = Type("ValidationResult", func() {
	Field(1, "valid", Boolean)
	Field(2, "user_id", String)
	Field(3, "email", String)
	Field(4, "reason", String)
	Required("valid")
})

var Credentials = Type("Credentials", func() {
	Field(1, "email", String, func() {
		Format(FormatEmail)
		Example("service@example.com")
	})
	Field(2, "password", String, func() {
		MinLength(8)
		Example("changeme123")
	})
	Required("email", "password")
})

var RegisterPayload = Type("RegisterPayload", func() {
	Extend(Credentials)
	Field(3, "display_name", String, func() {
		MinLength(3)
		Example("Service Admin")
	})
	Required("display_name")
})

var ValidateTokenPayload = Type("ValidateTokenPayload", func() {
	Field(1, "token", String, "JWT access token")
	Required("token")
})

var _ = Service("identity", func() {
	Description("Operations for user identities")

	Error("unauthorized", UnauthorizedError)
	Error("not_found", NotFoundError)

	Method("register", func() {
		Description("Registers a new user")
		Payload(RegisterPayload)
		Result(User)
		HTTP(func() {
			POST("/v1/identity/register")
			Response(StatusCreated)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("login", func() {
		Description("Authenticates a user and issues a JWT")
		Payload(Credentials)
		Result(TokenResult)
		HTTP(func() {
			POST("/v1/identity/login")
			Response(StatusOK)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("validate_token", func() {
		Description("Validates a JWT and returns the claims")
		Payload(ValidateTokenPayload)
		Result(ValidationResult)
		HTTP(func() {
			POST("/v1/identity/validate")
			Response(StatusOK)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Files("openapi.json", "gen/http/openapi.json")
})
