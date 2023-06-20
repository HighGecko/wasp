package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestIRC27NFT(t *testing.T) {
	testMIME := "fakeMIME"
	testURL := "http://no.org"
	testName := "hi-name"
	testNft := isc.NewIRC27NFTMetadata(testMIME, testURL, testName)
	data1, err := testNft.Bytes()
	require.NoError(t, err)
	nft, err := isc.IRC27NFTMetadataFromBytes(data1)
	require.NoError(t, err)
	require.Equal(t, testNft, nft)
	data2, err := nft.Bytes()
	require.NoError(t, err)
	require.Equal(t, data1, data2)
}
