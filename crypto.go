package main

import (
	"filippo.io/age"
	"io"
	"os"
	"strings"
)

func decryptFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	idReader := strings.NewReader("AGE-PLUGIN-YUBIKEY-1JAAZ6QVZ0RQLA0GD5ZEPL\n" +
		"AGE-PLUGIN-YUBIKEY-1ZJRRUQVZE9DJFCC3UPNJD\n" +
		"AGE-PLUGIN-YUBIKEY-1ZWRRUQVZAJX2Q4C5LJRTD")
	identities, err := age.ParseIdentities(idReader)
	if err != nil {
		return "", err
	}
	decrypt, err := age.Decrypt(file, identities...)
	if err != nil {
		return "", err
	}
	all, err := io.ReadAll(decrypt)
	if err != nil {
		return "", err
	}
	return string(all), nil
}
