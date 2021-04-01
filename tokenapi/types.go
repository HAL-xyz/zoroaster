package tokenapi

import "fmt"

type ApiNetworkErr struct {
	msg string
}

func (e ApiNetworkErr) Error() string {
	return fmt.Sprintf(e.msg)
}

type ApiNotFoundErr struct {
	msg string
}

func (e ApiNotFoundErr) Error() string {
	return fmt.Sprintf(e.msg)
}

type GeckoIDSJson []struct {
	ID        string `json:"id"`
	Platforms struct {
		Ethereum string `json:"ethereum"`
	} `json:"platforms"`
}

type ERC20Token struct {
	ChainId  int    `json:"chainId"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoURI,omitempty"`
}
