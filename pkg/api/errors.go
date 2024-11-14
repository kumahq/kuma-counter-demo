package api

type ErrorType struct {
	Key         string
	Description string
}

func (t ErrorType) QualifiedType() string {
	return "https://github.com/kumahq/kuma-counter-demo/blob/main/ERRORS.md#" + t.Key
}

var (
	INVALID_JSON   = ErrorType{Key: "INVALID-JSON", Description: "Invalid JSON"}
	INTERNAL_ERROR = ErrorType{Key: "INTERNAL-ERROR", Description: "A complex error occured"}
	KV_NOT_FOUND   = ErrorType{Key: "KV-NOT-FOUND", Description: "Couldn't find a kv entry"}
	KV_DISABLED    = ErrorType{Key: "KV-DISABLED", Description: "You can't use KV or KVUrl is not set. Are you talking to the right service?"}
	KV_CONFLICT    = ErrorType{Key: "KV-CONFLICT", Description: "A conflict when update a key with compare and swap"}
)
var ErrorTypes = []ErrorType{INVALID_JSON, INTERNAL_ERROR, KV_NOT_FOUND, KV_DISABLED, KV_CONFLICT}
