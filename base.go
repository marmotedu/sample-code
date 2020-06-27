package code

import (
	"net/http"
)

// 通用: 基本错误
// Code must start with 1xxxxx
const (
	// Success - 100001: No error occurred.
	ErrSuccess int = iota + 100001

	// ErrUnknown - 100002: Internal server error.
	ErrUnknown

	// ErrBind - 100003: Error occurred while binding the request body to the struct
	ErrBind

	// ErrValidation - 100004: Validation failed.
	ErrValidation

	// ErrTokenInvalid - 100005: Token invalid.
	ErrTokenInvalid
)

// 通用：数据库类错误
const (
	// ErrDatabase - 100101: Database error.
	ErrDatabase int = iota + 100101
)

// 通用：认证授权类错误
const (
	// ErrEncrypt - 100201: Error occurred while encrypting the user password.
	ErrEncrypt int = iota + 100201

	// ErrSignatureInvalid - 100202: Signature is invalid.
	ErrSignatureInvalid

	// ErrExpired - 100203: Token expired.
	ErrExpired

	// ErrInvalidAuthHeader - 100204: Invalid authorization header.
	ErrInvalidAuthHeader

	// ErrMissingHeader - 100205: The `Authorization` header was empty.
	ErrMissingHeader

	// ErrorExpired - 100206: Token expired.
	ErrorExpired

	// ErrPasswordIncorrect - 100207: Password was incorrect.
	ErrPasswordIncorrect
)

// 通用：编解码类错误
const (
	// ErrEncodingFailed - 100301: Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301

	// ErrDecodingFailed - 100302: Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON - 100303: Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON - 100304: JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON - 100305: JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml - 100306: Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml - 100307: Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml - 100308: Yaml data could not be decoded.
	ErrDecodingYaml
)

func init() {

	register(ErrSuccess, http.StatusOK, "")
	register(ErrUnknown, http.StatusInternalServerError, "Internal server error")
	register(ErrBind, http.StatusBadRequest, "Error occurred while binding the request body to the struct")
	register(ErrValidation, http.StatusBadRequest, "Validation failed")
	register(ErrTokenInvalid, http.StatusForbidden, "Token invalid")

	register(ErrDatabase, http.StatusInternalServerError, "Database error")

	register(ErrEncrypt, http.StatusUnauthorized, "Error occurred while encrypting the user password")
	register(ErrSignatureInvalid, http.StatusUnauthorized, "Signature is invalid")
	register(ErrorExpired, http.StatusUnauthorized, "Token expired")
	register(ErrInvalidAuthHeader, http.StatusUnauthorized, "Invalid authorization header")
	register(ErrMissingHeader, http.StatusUnauthorized, "The length of the `Authorization` header is zero.")
	register(ErrPasswordIncorrect, http.StatusUnauthorized, "Password was incorrect")

	register(ErrEncodingFailed, http.StatusInternalServerError, "Encoding failed due to an error with the data")
	register(ErrDecodingFailed, http.StatusInternalServerError, "Decoding failed due to an error with the data")
	register(ErrInvalidJSON, http.StatusInternalServerError, "Data is not valid JSON")
	register(ErrEncodingJSON, http.StatusInternalServerError, "JSON data could not be encoded")
	register(ErrDecodingJSON, http.StatusInternalServerError, "JSON data could not be decoded")
	register(ErrInvalidYaml, http.StatusInternalServerError, "Data is not valid Yaml")
	register(ErrEncodingYaml, http.StatusInternalServerError, "YAML data could not be encoded")
	register(ErrDecodingYaml, http.StatusInternalServerError, "YAML data could not be decoded")
}
