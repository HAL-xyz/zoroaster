package action

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/leekchan/accounting"
	"html/template"
	"math/big"
	"strconv"
	"strings"
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
		"round":                utils.Round,
		"pow":                  pow,
		"formatNumber":         formatNumber,
		"toFiat":               tokenapi.GetTokenAPI().GetExchangeRate,
		"floatToInt":           floatToInt,
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

func floatToInt(s string) int64 {
	n, _ := strconv.ParseFloat(s, 64)
	return int64(n)
}
