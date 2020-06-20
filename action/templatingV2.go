package action

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"html/template"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func renderTemplateWithData(templateText string, data interface{}) (string, error) {

	funcMap := template.FuncMap{
		"upperCase":            strings.ToUpper,
		"hexToASCII":           hexToASCII,
		"hexToInt":             hexToInt,
		"etherscanTxLink":      etherscanTxLink,
		"etherscanAddressLink": etherscanAddressLink,
		"etherscanTokenLink":   etherscanTokenLink,
		"fromWei":              fromWei,
		"humanTime":            timestampToHumanTime,
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
	return string(bs)
}

func hexToInt(s string) string {
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

func fromWei(wei interface{}, units int) string {
	switch v := wei.(type) {
	case *big.Int:
		return scaleBy(v.String(), fmt.Sprintf("%f", math.Pow10(units)))
	case string:
		return scaleBy(v, fmt.Sprintf("%f", math.Pow10(units)))
	case int:
		return scaleBy(strconv.Itoa(v), fmt.Sprintf("%f", math.Pow10(units)))
	default:
		return fmt.Sprintf("cannot use %s of type %T as wei input", wei, wei)
	}
}

func timestampToHumanTime(timestamp interface{}) string {
	switch v := timestamp.(type) {
	case int:
		unixTimeUTC := time.Unix(int64(v), 0)
		return unixTimeUTC.Format(time.RFC822)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return v
		}
		unixTimeUTC := time.Unix(i, 0)
		return unixTimeUTC.Format(time.RFC822)
	default:
		return fmt.Sprintf("cannot use %s of type %T as timestamp input", timestamp, timestamp)
	}
}
