package hubbleutils

var (
	ChainId           int64
	VerifyingContract string
)

func SetChainIdAndVerifyingSignedOrdersContract(chainId int64, verifyingContract string) {
	ChainId = chainId
	VerifyingContract = verifyingContract
}
