// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func hexDecrypt(s string) []byte {
	h, _ := hex.DecodeString(s)
	return h
}

var (
	emptyTestKey = make([]byte, aes.BlockSize)
	emptyIV      = make([]byte, aes.BlockSize)
)

// These are default values for error-less en- and decryption.
// The correct error handling will be done with different data sets.
var testCCvalues = map[string]struct {
	ccOdceType                  odceCipher
	symmetricKey, iv, plaintext string
	ciphertext                  []byte
}{
	"simple": {
		ccOdceType:   aes_256_cbc,
		symmetricKey: "not so secret key with more pad ",
		iv:           "non-random ivpad",
		plaintext:    "this is the desired content",
		ciphertext:   hexDecrypt("f72fdd97a7f9b35fcc04e1ca948f0a6ac67a130d1ca871f07db58fb8cf96fac5"),
	},
}

func TestCipherChannel_encrypt(t *testing.T) {
	for title, cas := range testCCvalues {
		t.Run(title,
			func(t *testing.T) {
				cc, err := newCipherChannel(cas.ccOdceType, []byte(cas.symmetricKey))
				if err != nil {
					t.Errorf("Error creating cipherChannel: %w", err)
					return
				}

				ciphertext, _, err := cc.encrypt([]byte(cas.plaintext), []byte(cas.iv))
				if err != nil {
					t.Errorf("Error during encryption: %w", err)
					return
				}

				if !bytes.Equal(cas.ciphertext, ciphertext) {
					t.Errorf("Received ciphertext unequal to expected ciphertext")
					t.Errorf("Expected: %v", cas.ciphertext)
					t.Errorf("Received: %v", ciphertext)
				}
			},
		)
	}
}

func TestCipherChannel_decrypt(t *testing.T) {
	for title, cas := range testCCvalues {
		t.Run(title,
			func(t *testing.T) {
				cc, err := newCipherChannel(cas.ccOdceType, []byte(cas.symmetricKey))
				if err != nil {
					t.Errorf("Error creating cipherChannel: %w", err)
					return
				}

				plaintext, err := cc.decrypt(cas.ciphertext, []byte(cas.iv))
				if err != nil {
					t.Errorf("Error during encryption: %w", err)
					return
				}

				// Trim trailing nullbytes - these are added due to the
				// padding.
				trimPlaintext := string(bytes.TrimRight(plaintext, "\x00"))

				if strings.Compare(cas.plaintext, trimPlaintext) != 0 {
					t.Errorf("Received plaintext unequal to expected plaintext")
					t.Errorf("Expected: %v", cas.plaintext)
					t.Errorf("Received: %v", trimPlaintext)
				}
			},
		)
	}
}

func TestCipherChannel_encrypt_Error_IVSize(t *testing.T) {
	cc, _ := newCipherChannel(aes_256_cbc, emptyTestKey)
	_, _, err := cc.encrypt(nil, []byte{0x0})
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}

	expectedErrString := fmt.Sprintf("passed IV is 1 bytes long instead of %d", aes.BlockSize)
	if err.Error() != expectedErrString {
		t.Errorf("Received error with invalid text:")
		t.Errorf("Expected: %s", expectedErrString)
		t.Errorf("Received: %s", err)
	}
}

func TestCipherChannel_decrypt_IVNil(t *testing.T) {
	cc, _ := newCipherChannel(aes_256_cbc, emptyTestKey)
	_, err := cc.decrypt(nil, nil)
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	expectedErrString := "passed IV is nil"
	if err.Error() != expectedErrString {
		t.Errorf("Received error with invalid text:")
		t.Errorf("Expected: %s", expectedErrString)
		t.Errorf("Received: %s", err)
	}
}

func TestCipherChannel_decrypt_Error_IVSize(t *testing.T) {
	cc, _ := newCipherChannel(aes_256_cbc, emptyTestKey)
	_, err := cc.decrypt(nil, []byte{0x0})
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}

	expectedErrString := fmt.Sprintf("passed IV is 1 bytes long instead of %d", aes.BlockSize)
	if err.Error() != expectedErrString {
		t.Errorf("Received error with invalid text:")
		t.Errorf("Expected: %s", expectedErrString)
		t.Errorf("Received: %s", err)
	}
}

func TestCipherChannel_decrypt_Error_CipherTextSize(t *testing.T) {
	cc, _ := newCipherChannel(aes_256_cbc, emptyTestKey)
	_, err := cc.decrypt([]byte{0x0}, emptyIV)
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}

	expectedErrString := fmt.Sprintf("passed ciphertext cannot be split into blocks of %d bytes", aes.BlockSize)
	if err.Error() != expectedErrString {
		t.Errorf("Received error with invalid text:")
		t.Errorf("Expected: %s", expectedErrString)
		t.Errorf("Received: %s", err)
	}
}
