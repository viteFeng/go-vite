package wallet_test

import (
	"github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/wallet"
	"path/filepath"
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestManager_NewMnemonicAndSeedStore(t *testing.T) {
	manager := wallet.New(&wallet.Config{
		DataDir: filepath.Join(common.DefaultDataDir(), "wallet"),
	})
	mnemonic, seedStoreFile, err := manager.NewMnemonicAndEntropyStore("123456", true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(mnemonic)
	fmt.Println(seedStoreFile)

	file, e := manager.RecoverEntropyStoreFromMnemonic(mnemonic, "123456", true)
	if e != nil {
		t.Fatal()
	}
	assert.Equal(t, file, seedStoreFile)
}
