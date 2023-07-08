package utils

import "github.com/ethereum/go-ethereum/log"

func GetLogMethod(err error) func(string, ...interface{}) {
	if err != nil {
		return log.Error
	}
	return log.Info
}
