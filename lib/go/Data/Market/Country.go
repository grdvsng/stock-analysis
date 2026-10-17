package Market

import "strings"

var exchangeCountries = map[string]string{
	"NASDAQ":    "US",
	"NYSE":      "US",
	"ARCA":      "US",
	"AMEX":      "US",
	"BATS":      "US",
	"IEX":       "US",
	"PINK":      "US",
	"ISLAND":    "US",
	"TSE":       "CA",
	"VENTURE":   "CA",
	"MEXI":      "MX",
	"LSE":       "GB",
	"LSEETF":    "GB",
	"IBIS":      "DE",
	"IBIS2":     "DE",
	"FWB":       "DE",
	"FWB2":      "DE",
	"SWB":       "DE",
	"GETTEX":    "DE",
	"SBF":       "FR",
	"AEB":       "NL",
	"ENEXT.BE":  "BE",
	"BVL":       "PT",
	"BM":        "ES",
	"BVME":      "IT",
	"EBS":       "CH",
	"VSE":       "AT",
	"SFB":       "SE",
	"CPH":       "DK",
	"OSE":       "NO",
	"HEX":       "FI",
	"WSE":       "PL",
	"PRA":       "CZ",
	"TASE":      "IL",
	"TSEJ":      "JP",
	"SEHK":      "HK",
	"SEHKNTL":   "CN",
	"SEHKSZSE":  "CN",
	"SGX":       "SG",
	"ASX":       "AU",
	"NSE":       "IN",
	"KSE":       "KR",
	"TWSE":      "TW",
	"JSE":       "ZA",
	"BOVESPA":   "BR",
	"CHINEXT":   "CN",
	"SHMSE":     "CN",
	"SZMSE":     "CN",
	"NZE":       "NZ",
	"OTCLNKECN": "US",
}

func CountryOf(primaryExchange string, isin string) string {
	if len(isin) >= 2 {
		return strings.ToUpper(isin[:2])
	}

	return exchangeCountries[strings.ToUpper(primaryExchange)]
}
