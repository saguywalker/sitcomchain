package code

const (
	CodeTypeOK = iota
	CodeTypeEncodingError
	CodeTypeBadNonce
	CodeTypeUnauthorized
	CodeTypeUnknownError
	CodeTypeDuplicateKey
	CodeTypeUnmarshalError
	CodeTypeDecodingError
	CodeTypeDuplicateNonce
	CodeTypeEmptyMethod
	CodeTypeInvalidMethod
)
