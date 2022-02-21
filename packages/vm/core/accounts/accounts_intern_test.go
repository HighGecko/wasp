package accounts

import (
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/marshalutil"

	"github.com/iotaledger/wasp/packages/kv"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Logf("Name: %s", Contract.Name)
	t.Logf("Description: %s", Contract.Description)
	t.Logf("Program hash: %s", Contract.ProgramHash.String())
	t.Logf("Hname: %s", Contract.Hname())
}

var dummyAssetID = [iotago.NativeTokenIDLength]byte{1, 2, 3}

func checkLedgerT(t *testing.T, state dict.Dict, cp string) *iscp.Assets {
	total := GetTotalL2Assets(state)
	// t.Logf("checkpoint '%s.%s':\n%s", curTest, cp, total.String())
	require.NotPanics(t, func() {
		checkLedger(state, cp)
	})
	return total
}

func TestCreditDebit1(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")

	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.KnownAgentID(1, 2)
	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	require.NotNil(t, total)
	require.EqualValues(t, 1, len(total.Tokens))
	require.True(t, total.Equals(transfer))

	transfer.Iotas = 1
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp2")

	expected := iscp.NewAssets(43, nil).AddNativeTokens(dummyAssetID, big.NewInt(4))
	require.True(t, expected.Equals(total))

	userAssets := GetAssets(state, agentID1)
	require.EqualValues(t, 43, userAssets.Iotas)
	require.Zero(t, userAssets.Tokens.MustSet()[dummyAssetID].Amount.Cmp(big.NewInt(4)))
	checkLedgerT(t, state, "cp2")

	DebitFromAccount(state, agentID1, expected)
	total = checkLedgerT(t, state, "cp3")
	expected = iscp.NewEmptyAssets()
	require.True(t, expected.Equals(total))
}

func TestCreditDebit2(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.NewRandomAgentID()
	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.Tokens))
	require.True(t, expected.Equals(total))

	transfer = iscp.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	DebitFromAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp2")
	require.EqualValues(t, 0, len(total.Tokens))
	expected = iscp.NewAssets(42, nil)
	require.True(t, expected.Equals(total))

	require.True(t, util.IsZeroBigInt(GetNativeTokenBalance(state, agentID1, &transfer.Tokens[0].ID)))
	bal1 := GetAccountAssets(state, agentID1)
	require.False(t, bal1.IsEmpty())
	require.True(t, total.Equals(bal1))
}

func TestCreditDebit3(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.NewRandomAgentID()
	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.Tokens))
	require.True(t, expected.Equals(total))

	transfer = iscp.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(100))
	require.Panics(t,
		func() {
			DebitFromAccount(state, agentID1, transfer)
		},
	)
	total = checkLedgerT(t, state, "cp2")

	require.EqualValues(t, 1, len(total.Tokens))
	expected = iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))
}

func TestCreditDebit4(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.NewRandomAgentID()
	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.Tokens))
	require.True(t, expected.Equals(total))

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = iscp.NewAssets(20, nil)
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedgerT(t, state, "cp2")

	keys = getAccountsIntern(state).Keys()
	require.EqualValues(t, 2, len(keys))

	expected = iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := GetAccountAssets(state, agentID1)
	require.False(t, bm1.IsEmpty())
	expected = iscp.NewAssets(22, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(bm1))

	bm2 := GetAccountAssets(state, agentID2)
	require.False(t, bm2.IsEmpty())
	expected = iscp.NewAssets(20, nil)
	require.True(t, expected.Equals(bm2))
}

func TestCreditDebit5(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.NewRandomAgentID()
	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	total = checkLedgerT(t, state, "cp1")

	expected := transfer
	require.EqualValues(t, 1, len(total.Tokens))
	require.True(t, expected.Equals(total))

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	transfer = iscp.NewAssets(50, nil)
	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.False(t, ok)
	total = checkLedgerT(t, state, "cp2")

	keys = getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	expected = iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	require.True(t, expected.Equals(total))

	bm1 := GetAccountAssets(state, agentID1)
	require.False(t, bm1.IsEmpty())
	require.True(t, expected.Equals(bm1))

	bm2 := GetAccountAssets(state, agentID2)
	require.True(t, bm2.IsEmpty())
}

func TestCreditDebit6(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.NewRandomAgentID()
	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	checkLedgerT(t, state, "cp1")

	agentID2 := iscp.NewRandomAgentID()
	require.NotEqualValues(t, agentID1, agentID2)

	ok := MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.True(t, ok)
	total = checkLedgerT(t, state, "cp2")

	keys := getAccountsIntern(state).Keys()
	require.EqualValues(t, 1, len(keys))

	bal := GetAccountAssets(state, agentID1)
	require.True(t, bal.IsEmpty())

	bal2 := GetAccountAssets(state, agentID2)
	require.False(t, bal2.IsEmpty())
	require.True(t, total.Equals(bal2))
}

func TestCreditDebit7(t *testing.T) {
	state := dict.New()
	total := checkLedgerT(t, state, "cp0")
	require.True(t, total.Equals(iscp.NewEmptyAssets()))

	agentID1 := iscp.NewRandomAgentID()
	transfer := iscp.NewEmptyAssets().AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	checkLedgerT(t, state, "cp1")

	debitTransfer := iscp.NewAssets(1, nil)
	// debit must fail
	require.Panics(t, func() {
		DebitFromAccount(state, agentID1, debitTransfer)
	})

	total = checkLedgerT(t, state, "cp1")
	require.True(t, transfer.Equals(total))
}

func TestMoveAll(t *testing.T) {
	state := dict.New()
	agentID1 := iscp.NewRandomAgentID()
	agentID2 := iscp.NewRandomAgentID()

	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	require.EqualValues(t, 1, getAccountsMapR(state).MustLen())
	accs := getAccountsIntern(state)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	MoveBetweenAccounts(state, agentID1, agentID2, transfer)
	require.EqualValues(t, 1, getAccountsMapR(state).MustLen())
	accs = getAccountsIntern(state)
	require.EqualValues(t, 1, len(accs))
	_, ok = accs[kv.Key(agentID2.Bytes())]
	require.True(t, ok)
}

func TestDebitAll(t *testing.T) {
	state := dict.New()
	agentID1 := iscp.NewRandomAgentID()

	transfer := iscp.NewAssets(42, nil).AddNativeTokens(dummyAssetID, big.NewInt(2))
	CreditToAccount(state, agentID1, transfer)
	require.EqualValues(t, 1, getAccountsMapR(state).MustLen())
	accs := getAccountsIntern(state)
	require.EqualValues(t, 1, len(accs))
	_, ok := accs[kv.Key(agentID1.Bytes())]
	require.True(t, ok)

	DebitFromAccount(state, agentID1, transfer)
	require.EqualValues(t, 0, getAccountsMapR(state).MustLen())
	accs = getAccountsIntern(state)
	require.EqualValues(t, 0, len(accs))
	require.True(t, ok)

	assets := GetAssets(state, agentID1)
	require.True(t, assets.IsEmpty())

	assets = GetTotalL2Assets(state)
	require.True(t, assets.IsEmpty())
}

func TestFoundryOutputRec(t *testing.T) {
	o := foundryOutputRec{
		Amount:            300,
		TokenTag:          iotago.TokenTag{},
		TokenScheme:       &iotago.SimpleTokenScheme{},
		MaximumSupply:     big.NewInt(1000),
		CirculatingSupply: big.NewInt(20),
		BlockIndex:        3,
		OutputIndex:       2,
	}
	oBin := o.Bytes()
	o1, err := foundryOutputRecFromMarshalUtil(marshalutil.New(oBin))
	require.NoError(t, err)
	require.EqualValues(t, o.Amount, o1.Amount)
	require.EqualValues(t, o.TokenTag, o1.TokenTag)
	_, ok := o1.TokenScheme.(*iotago.SimpleTokenScheme)
	require.True(t, ok)
	require.True(t, o.MaximumSupply.Cmp(o1.MaximumSupply) == 0)
	require.True(t, o.CirculatingSupply.Cmp(o1.CirculatingSupply) == 0)
	require.EqualValues(t, o.BlockIndex, o1.BlockIndex)
	require.EqualValues(t, o.OutputIndex, o1.OutputIndex)
}

func TestCreditDebitNFT1(t *testing.T) {
	state := dict.New()

	agentID1 := iscp.KnownAgentID(1, 2)
	nft := iotago.NFTID{123}
	CreditNFTToAccount(state, agentID1, &nft)

	accNFTs := GetAccountNFTs(state, agentID1)
	require.Len(t, accNFTs, 1)
	require.Equal(t, accNFTs[0], nft)

	DebitNFTFromAccount(state, agentID1, &nft)

	accNFTs = GetAccountNFTs(state, agentID1)
	require.Len(t, accNFTs, 0)
}
