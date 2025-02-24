package wallet

import "errors"

type walletErrors struct {
	InsufficientFunds error
	RecipientNotFound error
}

func errorFactory() walletErrors {
	return walletErrors{
		//nolint:err113 // false positive
		InsufficientFunds: errors.New("insufficient funds"),
		//nolint:err113 // false positive
		RecipientNotFound: errors.New("recipient not found"),
	}
}
