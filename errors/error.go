package errors

const (
	//300
	RESOURCE_ALREADY_EXISTS string = "302"
	RESOURCE_NOT_FOUND      string = "301"
	RESOURCE_EXHAUSTED      string = "303"
	//400
	SERVICE_INTERNAL  string = "402"
	SERVICE_VIOLATION string = "406"
	//600
	SERVICE_UNAVAILABLE string = "601"

	// messages
	// 300
	RESOURCE_ALREADY_EXISTS_MSG string = "Failed to parse payload"
	RESOURCE_NOT_FOUND_MSG      string = "Failed to parse payload"
	RESOURCE_EXHAUSTED_MSG      string = "303"
	// 400
	SERVICE_INTERNAL_MSG  string = "Failed to parse payload"
	SERVICE_VIOLATION_MSG string = ""
	// 600
	SERVICE_UNAVAILABLE_MSG string = "601"
)
