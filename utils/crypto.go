package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"go.uber.org/zap"
	zlog "walletSynV2/utils/zlog_sing"
)

func RedPacketCrypto(str string) (string, error) {
	plaintext := []byte(str)
	keyStr := "4+sorWFPdy0jwpU/"
	key := []byte(keyStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		zlog.Zlog.Error("Str crypto error", zap.Error(err))
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
