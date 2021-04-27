package action

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/leekchan/accounting"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func RenderTemplateWithData(templateText string, data interface{}) (string, error) {

	funcMap := template.FuncMap{
		"upperCase":            strings.ToUpper,
		"hexToASCII":           hexToASCII,
		"hexToInt":             hexToInt,
		"etherscanTxLink":      etherscanTxLink,
		"etherscanAddressLink": etherscanAddressLink,
		"etherscanTokenLink":   etherscanTokenLink,
		"fromWei":              tokenapi.GetTokenAPI().FromWei,
		"humanTime":            timestampToHumanTime,
		"symbol":               tokenapi.GetTokenAPI().Symbol,
		"decimals":             tokenapi.GetTokenAPI().Decimals,
		"balanceOf":            tokenapi.GetTokenAPI().BalanceOf,
		"add":                  add,
		"sub":                  sub,
		"mul":                  mul,
		"div":                  div,
		"percentageVariation":  percentageVariation,
		"round":                utils.Round,
		"pow":                  pow,
		"formatNumber":         formatNumber,
		"toFiat":               wrapGetExchangeRate,
		"toFiatAt":             wrapGetExchangeRateAtDate,
		"floatToInt":           floatToInt,
		"ERC20Snapshot":        eRC20Snapshot,
		"ethCall":              ethCall,
	}

	tmpl := template.New("").Funcs(funcMap)
	t, err := tmpl.Parse(templateText)

	if err != nil {
		return fmt.Sprintf("Could not parse template: %s\n", err), err
	}

	var output bytes.Buffer
	err = t.Execute(&output, data)

	return output.String(), err
}

func hexToASCII(s string) string {
	s = strings.TrimPrefix(s, "0x")
	bs, err := hex.DecodeString(s)
	if err != nil {
		return s
	}
	bs = bytes.Trim(bs, "\x00")
	return string(bs)
}

func hexToInt(s string) string {
	if !strings.HasPrefix(s, "0x") {
		return s
	}
	s = strings.TrimPrefix(s, "0x")
	i := new(big.Int)
	_, ok := i.SetString(s, 16)
	if !ok {
		return s
	}
	return i.String()
}

func etherscanTxLink(hash string) string {
	return fmt.Sprintf("https://etherscan.io/tx/%s", hash)
}

func etherscanAddressLink(address string) string {
	return fmt.Sprintf("https://etherscan.io/address/%s", address)
}

func etherscanTokenLink(token string) string {
	return fmt.Sprintf("https://etherscan.io/token/%s", token)
}

func timestampToHumanTime(timestamp interface{}, optionalFormatting ...string) string {
	var format string
	if len(optionalFormatting) == 1 {
		format = optionalFormatting[0]
	} else {
		format = time.RFC822
	}
	switch v := timestamp.(type) {
	case int:
		unixTimeUTC := time.Unix(int64(v), 0).UTC()
		return unixTimeUTC.Format(format)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return v
		}
		unixTimeUTC := time.Unix(i, 0).UTC()
		return unixTimeUTC.Format(format)
	default:
		return fmt.Sprintf("cannot use %s of type %T as timestamp input", timestamp, timestamp)
	}
}

func add(nums ...interface{}) *big.Float {
	total := big.NewFloat(0)
	for _, i := range nums {
		total = total.Add(total, utils.MakeBigFloat(i))
	}
	return total
}

func sub(a, b interface{}) *big.Float {
	total := utils.MakeBigFloat(a)
	return total.Sub(total, utils.MakeBigFloat(b))
}

func mul(a, b interface{}) *big.Float {
	x := utils.MakeBigFloat(a)
	y := utils.MakeBigFloat(b)
	return x.Mul(x, y)
}

func percentageVariation(new, old interface{}) string {
	diff := sub(new, old)
	variation := mul(div(diff, old), 100)
	//sign := ""
	//if variation.Sign() == 1 {
	//	sign = "+"
	//}
	return fmt.Sprintf("%s%%", formatNumber(variation.String(), 2))
}

func div(a, b interface{}) *big.Float {
	x := utils.MakeBigFloat(a)
	y := utils.MakeBigFloat(b)
	return new(big.Float).Quo(x, y)
}

func pow(a, b interface{}) *big.Float {
	x := utils.MakeBigFloat(a)

	var y int
	if v, ok := b.(int); ok {
		y = v
	}
	if v, err := strconv.Atoi(fmt.Sprintf("%v", b)); err == nil {
		y = v
	}

	var result = big.NewFloat(0.0).Copy(x)
	for i := 0; i < y-1; i++ {
		result = big.NewFloat(0.0).SetPrec(256).Mul(result, x)
	}
	return result
}

func formatNumber(number interface{}, precision int) string {
	return accounting.FormatNumberBigFloat(utils.MakeBigFloat(number), precision, ",", ".")
}

func floatToInt(i interface{}) int64 {
	switch v := i.(type) {
	case string:
		n, _ := strconv.ParseFloat(v, 64)
		return int64(n)
	case *big.Float:
		return floatToInt(v.String())
	default:
		return 0
	}
}

func eRC20Snapshot(allBalancesIfc []interface{}) map[string]*big.Int {
	// balances are already sorted per address because the multicall is ordered;
	// here we are just converting the multicall output to []*big.Int

	if len(allBalancesIfc) != 1 {
		return map[string]*big.Int{}
	}
	balances, ok := allBalancesIfc[0].([]string)
	if !ok {
		return map[string]*big.Int{}
	}

	sortedBalances := make([]*big.Int, len(balances))
	for i, v := range balances {
		sortedBalances[i] = utils.MakeBigInt(v)
	}

	// make a sorted list of all the tokens
	var i = 0
	sortedTokenAdds := make([]string, len(tokenapi.GetTokenAPI().GetAllERC20TokensMap()))
	for k := range tokenapi.GetTokenAPI().GetAllERC20TokensMap() {
		sortedTokenAdds[i] = k
		i++
	}
	sort.Strings(sortedTokenAdds)

	// So now we have:
	// an [] of all Balances Sorted
	// an [] of all Tokens Sorted

	// combine []balances and []tokenAdds to a balance map (tokenAdd -> balance)
	var balanceMap = map[string]*big.Int{}

	for i, balance := range sortedBalances {
		if balance.Cmp(big.NewInt(0)) == 1 {
			balanceMap[sortedTokenAdds[i]] = balance
		}
	}

	return balanceMap
}

// The template system doesn't like functions that return (T, error);
// in fact, it will abort parsing the template altogether.
// So we're wrapping the original functions to provide a dummy exchange value in case of errors;
// this way the result won't make sense, but at least it won't break everything.
func wrapGetExchangeRate(tokenAddress, fiatCurrency string) float32 {
	res, err := tokenapi.GetTokenAPI().GetExchangeRate(tokenAddress, fiatCurrency)
	if err != nil {
		return 0
	}
	return res
}

func wrapGetExchangeRateAtDate(tokenAddress, fiatCurrency, when string) float32 {
	res, err := tokenapi.GetTokenAPI().GetExchangeRateAtDate(tokenAddress, fiatCurrency, when)
	if err != nil {
		return 0
	}
	return res
}

func ethCall(address string, blockNo, returnedPosition int, method string, args ...string) string {

	res, err := tokenapi.GetTokenAPI().EthCall(address, method, "", blockNo, args...)
	if err != nil {
		return err.Error()
	}
	if returnedPosition >= len(res) {
		return fmt.Sprintf("invalid returned position %d", returnedPosition)
	}

	printableResults := utils.SprintfInterfaces(res)
	return fmt.Sprintf("%v", printableResults[returnedPosition])
}
