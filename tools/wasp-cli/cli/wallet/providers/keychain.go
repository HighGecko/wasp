package providers

import (
	"reflect"

	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type KeyChainWallet struct {
	cryptolib.VariantKeyPair
	addressIndex uint32
}

func newInMemoryWallet(keyPair *cryptolib.KeyPair, addressIndex uint32) *KeyChainWallet {
	return &KeyChainWallet{
		VariantKeyPair: keyPair,
		addressIndex:   addressIndex,
	}
}

func (i *KeyChainWallet) AddressIndex() uint32 {
	return i.addressIndex
}

func LoadKeyChain(addressIndex uint32) wallets.Wallet {
	seed, err := config.GetKeyChain().GetSeed()
	log.Check(err)

	useLegacyDerivation := config.GetUseLegacyDerivation()
	keyPair := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(seed[:], addressIndex, useLegacyDerivation))

	return newInMemoryWallet(keyPair, addressIndex)
}

func CreateKeyChain(overwrite bool) {
	oldSeed, _ := config.GetKeyChain().GetSeed()

	if len(oldSeed) == cryptolib.SeedSize && !overwrite {
		log.Printf("You already have an existing seed inside your Keychain.\nCalling `init` will *replace* it with a new one.\n")
		log.Printf("Run `wasp-cli init --overwrite` to continue with the initialization.\n")
		log.Fatalf("The cli will now exit.")
	}

	seed := cryptolib.NewSeed()
	err := config.GetKeyChain().SetSeed(seed)
	log.Check(err)

	log.Printf("New seed stored inside the Keychain.\n")
}

func MigrateKeyChain(seed cryptolib.Seed) {
	err := config.GetKeyChain().SetSeed(seed)
	log.Check(err)
	log.Printf("Seed migrated to Keychain.\nProceeding seed validation.\n")

	kcSeed, err := config.GetKeyChain().GetSeed()
	log.Check(err)

	if reflect.DeepEqual(kcSeed[:], seed[:]) {
		log.Printf("Seed has been successfully validated.\n")
		config.RemoveSeedForMigration()
		err = viper.WriteConfig()
		log.Check(err)
		log.Printf("Seed was removed from the wasp-cli.json\n")
	} else {
		log.Fatalf("Seed mismatch between Keychain and the wasp-cli.json.\nMigration failed.\n")

	}
}
