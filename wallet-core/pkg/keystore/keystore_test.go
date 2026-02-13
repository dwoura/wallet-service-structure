package keystore

import (
	"os"
	"testing"
)

func TestEncryptDecryptMnemonic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "secure-password"

	// 1. Encrypt
	keyJSON, err := EncryptMnemonic(mnemonic, password)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if keyJSON.Crypto.Cipher != "aes-256-gcm" {
		t.Errorf("Expected cipher aes-256-gcm, got %s", keyJSON.Crypto.Cipher)
	}

	// 2. Decrypt with correct password
	plaintext, err := DecryptMnemonic(keyJSON, password)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if plaintext != mnemonic {
		t.Errorf("Decryption mismatch. Expected %s, got %s", mnemonic, plaintext)
	}

	// 3. Decrypt with wrong password
	_, err = DecryptMnemonic(keyJSON, "wrong-password")
	if err == nil {
		t.Error("Expected error with wrong password, got nil")
	}
}

func TestFileSaveLoad(t *testing.T) {
	mnemonic := "test mnemonic"
	password := "123456"
	filename := "test_wallet.json"

	defer os.Remove(filename)

	// Encrypt
	keyJSON, _ := EncryptMnemonic(mnemonic, password)

	// Save
	err := keyJSON.SaveToFile(filename)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Load
	loadedJSON, err := LoadFromFile(filename)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify
	if loadedJSON.Id != keyJSON.Id {
		t.Errorf("ID mismatch after load")
	}

	// Decrypt Loaded
	decrypted, err := DecryptMnemonic(loadedJSON, password)
	if err != nil {
		t.Fatalf("Decrypt loaded failed: %v", err)
	}
	if decrypted != mnemonic {
		t.Errorf("Content mismatch")
	}
}
