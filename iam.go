package code

import (
	"net/http"
)

// iam服务：用户类错误
const (
	// ErrUserNotFound - 110001: User not found.
	ErrUserNotFound int = iota + 110001

	// ErrUserAlreadyExist - 110002: User already exist.
	ErrUserAlreadyExist
)

// iam服务：密钥类错误
const (
	// ErrEncrypt - 110101: Secret reach the max count.
	ErrReachMaxCount int = iota + 110101

	//  ErrSecretNotFound - 110102: Secret not found.
	ErrSecretNotFound
)

func init() {
	register(ErrUserNotFound, http.StatusNotFound, "User not found")
	register(ErrUserAlreadyExist, http.StatusBadRequest, "User already exist")

	register(ErrReachMaxCount, http.StatusBadRequest, "Secret reach the max count")
	register(ErrSecretNotFound, http.StatusNotFound, "Secret not found")
}
