package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("dummy", func() {
	Title("Dummy Service")
	Description("Reference CRUD microservice that enforces identity auth")

	Server("dummy", func() {
		Host("localhost", func() {
			URI("http://localhost:8082")
			URI("grpc://localhost:9082")
		})
		Services("dummy")
	})
})

var Item = ResultType("application/vnd.dummy.item", func() {
	TypeName("Item")
	Attributes(func() {
		Field(1, "id", String, "Item identifier")
		Field(2, "name", String)
		Field(3, "description", String)
		Field(4, "owner_id", String)
		Field(5, "created_at", String, func() {
			Format(FormatDateTime)
		})
		Required("id", "name", "owner_id", "created_at")
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("description")
		Attribute("owner_id")
		Attribute("created_at")
	})
})

var ItemsCollection = Type("ItemsCollection", func() {
	Field(1, "items", ArrayOf(Item))
	Required("items")
})

var DummyUnauthorizedError = Type("DummyUnauthorizedError", func() {
	Field(1, "message", String)
	Required("message")
})

var DummyNotFoundError = Type("DummyNotFoundError", func() {
	Field(1, "message", String)
	Required("message")
})

var AuthenticatedPayload = Type("AuthenticatedPayload", func() {
	Field(1, "token", String, "Bearer token")
	Required("token")
})

var CreateItemPayload = Type("CreateItemPayload", func() {
	Extend(AuthenticatedPayload)
	Field(2, "name", String)
	Field(3, "description", String)
	Required("name")
})

var ItemIDPayload = Type("ItemIDPayload", func() {
	Extend(AuthenticatedPayload)
	Field(2, "id", String)
	Required("id")
})

var ListItemsPayload = Type("ListItemsPayload", func() {
	Extend(AuthenticatedPayload)
})

var _ = Service("dummy", func() {
	Description("CRUD operations on items that rely on identity-api for auth")

	Error("unauthorized", DummyUnauthorizedError)
	Error("not_found", DummyNotFoundError)

	Method("create_item", func() {
		Payload(CreateItemPayload)
		Result(Item)
		HTTP(func() {
			POST("/v1/dummy/items")
			Header("token:Authorization", String, "Bearer token")
			Response(StatusCreated)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("list_items", func() {
		Payload(ListItemsPayload)
		Result(ItemsCollection)
		HTTP(func() {
			GET("/v1/dummy/items")
			Header("token:Authorization", String, "Bearer token")
			Response(StatusOK)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("get_item", func() {
		Payload(ItemIDPayload)
		Result(Item)
		HTTP(func() {
			GET("/v1/dummy/items/{id}")
			Header("token:Authorization", String, "Bearer token")
			Response(StatusOK)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("delete_item", func() {
		Payload(ItemIDPayload)
		Result(Empty)
		HTTP(func() {
			DELETE("/v1/dummy/items/{id}")
			Header("token:Authorization", String, "Bearer token")
			Response(StatusNoContent)
		})
		GRPC(func() {
			Response(CodeOK)
		})
	})

	Files("openapi.json", "gen/http/openapi.json")
})
