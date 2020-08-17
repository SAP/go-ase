// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// rsaEncrypt encrypts a password with the given public key using the
// given nonce.
func rsaEncrypt(pemPubKey, nonce, password []byte) ([]byte, error) {
	pubKeyBlock, rest := pem.Decode(pemPubKey)
	if len(rest) > 0 {
		return nil, fmt.Errorf("trailing bytes in public key: %#v", rest)
	}

	publicKey, err := x509.ParsePKCS1PublicKey(pubKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#1 public key: %w", err)
	}

	return rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, append(nonce, []byte(password)...), []byte{})
}

// generateSymmetricKey creates a cryptographically secure key to use
// for the odceCipher.
func generateSymmetricKey(odce odceCipher) ([]byte, error) {
	keyByteLength := 0
	switch odce {
	case aes_256_cbc:
		keyByteLength = 32
	default:
		return nil, fmt.Errorf("unhandled on demand encryption type: %s", odce)
	}

	bs := make([]byte, keyByteLength)
	readBytes, err := rand.Read(bs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	if readBytes != len(bs) {
		return nil, fmt.Errorf("read %d random bytes for symmetric key, expected %d",
			readBytes, len(bs))
	}

	return bs, nil
}

//go:generate stringer -type=odceCipher
// OnDemand Command Encryption cipher
type odceCipher int

const (
	aes_256_cbc odceCipher = iota
)

// cipherChannel en- or decrypts bytes written to or read from it.
// It is used to en- and decrypt the data portion of packets used to
// communicate with TDS servers.
type cipherChannel struct {
	odce         odceCipher
	symmetricKey []byte
	aes          cipher.Block
}

func newCipherChannel(odce odceCipher, symmetricKey []byte) (*cipherChannel, error) {
	aesBlock, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}

	cipherChannel := &cipherChannel{
		odce:         odce,
		symmetricKey: symmetricKey,
		aes:          aesBlock,
	}

	return cipherChannel, nil
}

// encrypt encrypts plaintext with the passed IV and returns the
// ciphertext and the used IV.
// If the passed iv is nil a random IV is chosen.
func (cc cipherChannel) encrypt(plaintext, iv []byte) ([]byte, []byte, error) {
	if iv == nil {
		// iv is null, this is the first data sent in an encrypted
		// stream - generate a random iv.
		iv = make([]byte, aes.BlockSize)
		if _, err := rand.Read(iv); err != nil {
			return nil, nil, fmt.Errorf("error reading random bytes for IV: %w", err)
		}
	} else {
		if len(iv) != aes.BlockSize {
			return nil, nil, fmt.Errorf("passed IV is %d bytes long instead of %d", len(iv), aes.BlockSize)
		}
	}

	// Add padding to the next full 16 bytes
	padding := make([]byte, aes.BlockSize-(len(plaintext)%aes.BlockSize))
	// Create ciphertext large enough to hold the plaintext and the
	// padding
	ciphertext := make([]byte, len(plaintext)+len(padding))

	cbcenc := cipher.NewCBCEncrypter(cc.aes, iv)
	cbcenc.CryptBlocks(ciphertext, append(plaintext, padding...))

	return ciphertext, iv, nil
}

// decrypt decrypts ciphertext with the passed IV and returns the
// plaintext.
func (cc cipherChannel) decrypt(ciphertext, iv []byte) ([]byte, error) {
	if iv == nil {
		return nil, fmt.Errorf("passed IV is nil")
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("passed IV is %d bytes long instead of %d", len(iv), aes.BlockSize)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("passed ciphertext cannot be split into blocks of %d bytes", aes.BlockSize)
	}

	plaintext := make([]byte, len(ciphertext))

	cbcdec := cipher.NewCBCDecrypter(cc.aes, iv)
	cbcdec.CryptBlocks(plaintext, ciphertext)

	return plaintext, nil
}
