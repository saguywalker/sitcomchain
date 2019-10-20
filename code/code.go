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
)
