package storage

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/hex"
    "io"
)

func Encrypt(key, plaintext []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

    return hex.EncodeToString(ciphertext), nil
}

func Decrypt(key []byte, ciphertext string) ([]byte, error) {
    cipherBytes, err := hex.DecodeString(ciphertext)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    if len(cipherBytes) < aes.BlockSize {
        return nil, err
    }

    iv := cipherBytes[:aes.BlockSize]
    cipherBytes = cipherBytes[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(cipherBytes, cipherBytes)

    return cipherBytes, nil
}
