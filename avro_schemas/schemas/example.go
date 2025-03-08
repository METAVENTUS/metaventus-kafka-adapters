package schemas

const ExampleName = "ModelExample"
const ExampleSchema = `{
		"type": "record",
		"name": "UserCreated",
		"fields": [
			{"name": "id", "type": "string"},
			{"name": "email", "type": "string"},
			{"name": "name", "type": "string"}
		]
	}`
