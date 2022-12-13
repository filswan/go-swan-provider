package utils

import (
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/dgrijalva/jwt-go"

	"github.com/shopspring/decimal"
)

// GetEpochInMillis get current timestamp
func GetEpochInMillis() (millis int64) {
	nanos := time.Now().UnixNano()
	millis = nanos / 1000000
	return
}

func GetInt64FromStr(numStr string) int64 {
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}

	return num
}

func GetFloat64FromStr(numStr *string) (float64, error) {
	if numStr == nil || *numStr == "" {
		return -1, nil
	}

	*numStr = strings.Trim(*numStr, " ")
	if *numStr == "" {
		return -1, nil
	}

	num, err := strconv.ParseFloat(*numStr, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1, err
	}

	return num, nil
}

func GetIntFromStr(numStr string) (int, error) {
	num, err := strconv.ParseInt(numStr, 10, 32)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1, err
	}

	return int(num), nil
}

func GetNumStrFromStr(numStr string) string {
	re := regexp.MustCompile("[0-9]+.?[0-9]*")
	words := re.FindAllString(numStr, -1)
	//logs.GetLogger().Info("words:", words)
	if len(words) > 0 {
		return words[0]
	}

	return ""
}

func GetByteSizeFromStr(sizeStr string) int64 {
	sizeStr = strings.Trim(sizeStr, " ")
	numStr := GetNumStrFromStr(sizeStr)
	numStr = strings.Trim(numStr, " ")
	size, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}
	unit := strings.Trim(sizeStr, numStr)
	unit = strings.Trim(unit, " ")
	if len(unit) == 0 {
		return size
	}

	unit = strings.ToUpper(unit)
	switch unit {
	case "GIB", "GB":
		size = size * 1024 * 1024 * 1024
	case "MIB", "MB":
		size = size * 1024 * 1024
	case "KIB", "KB":
		size = size * 1024
	case "BYTE", "B":
		return size
	default:
		return -1
	}

	return size
}

func IsSameDay(nanoSec1, nanoSec2 int64) bool {
	year1, month1, day1 := time.Unix(0, nanoSec1).Date()
	year2, month2, day2 := time.Unix(0, nanoSec2).Date()

	if year1 == year2 && month1 == month2 && day1 == day2 {
		return true
	}

	return false
}

func GetRandInRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	randVal := min + rand.Intn(max-min+1)
	return randVal
}

func IsStrEmpty(str *string) bool {
	if str == nil || *str == "" {
		return true
	}

	strTrim := strings.Trim(*str, " ")
	return len(strTrim) == 0
}

func GetDayNumFromEpoch(epoch int) int {
	return epoch / 2 / 60 / 24
}

func GetEpochFromDay(day int) int {
	return day * 24 * 60 * 2
}

func GetMinFloat64(val1, val2 *float64) *float64 {
	if val1 == nil {
		return val2
	}

	if val2 == nil {
		return val1
	}

	if *val1 <= *val2 {
		return val1
	}

	return val2
}

func GetCurrentEpoch() int {
	currentNanoSec := time.Now().UnixNano()
	currentEpoch := (currentNanoSec/1e9 - 1598306471) / 30
	return int(currentEpoch)
}

func GetDecimalFromStr(source string) (*decimal.Decimal, error) {
	re := regexp.MustCompile("[0-9]+.?[0-9]*")
	words := re.FindAllString(source, -1)
	if len(words) > 0 {
		numStr := strings.Trim(words[0], " ")
		result, err := decimal.NewFromString(numStr)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		return &result, nil
	}

	return nil, nil
}

func UrlJoin(root string, parts ...string) string {
	url := root

	for _, part := range parts {
		url = strings.TrimRight(url, "/") + "/" + strings.TrimLeft(part, "/")
	}
	url = strings.TrimRight(url, "/")

	return url
}

func RandStringRunes(letterRunes []rune, strLen int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, strLen)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandString(optionChars string, strLen int) string {
	rand.Seed(time.Now().UnixNano())
	result := ""
	for i := 0; i < strLen; i++ {
		random := rand.Intn(len(optionChars))
		result = result + optionChars[random:random+1]
	}
	return result
}

func DecodeJwtToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, nil)
	if token == nil {
		return nil, err
	}
	claims, _ := token.Claims.(jwt.MapClaims)

	//for key, element := range claims {
	//	fmt.Println("Key:", key, "=>", "Element:", element)
	//}

	return claims, nil
}

// https://docs.filecoin.io/store/lotus/very-large-files/#maximizing-storage-per-sector
func CalculatePieceSize(fileSize int64) (int64, float64) {
	exp := math.Ceil(math.Log2(float64(fileSize)))
	sectorSize2Check := math.Pow(2, exp)
	pieceSize2Check := int64(sectorSize2Check * 254 / 256)
	if fileSize <= pieceSize2Check {
		return pieceSize2Check, sectorSize2Check
	}

	exp = exp + 1
	realSectorSize := math.Pow(2, exp)
	realPieceSize := int64(realSectorSize * 254 / 256)
	return realPieceSize, realSectorSize
}

func CalculateRealCost(sectorSizeBytes float64, pricePerGiB decimal.Decimal) decimal.Decimal {
	//logs.GetLogger().Info("sectorSizeBytes:", sectorSizeBytes, " pricePerGiB:", pricePerGiB)
	bytesPerGiB := decimal.NewFromInt(1024 * 1024 * 1024)
	sectorSizeGiB := decimal.NewFromFloat(sectorSizeBytes).Div(bytesPerGiB)
	realCost := sectorSizeGiB.Mul(pricePerGiB)
	//logs.GetLogger().Info("realCost:", realCost)
	return realCost
}

func Convert2Title(text string) string {
	result := ""
	separator := "."
	sentences := strings.Split(text, separator)
	for _, sentence := range sentences {
		sentence = strings.Trim(sentence, " ")
		if len(sentence) == 0 {
			continue
		}
		firstChar := byte(unicode.ToUpper(rune(sentence[0])))
		if result != "" {
			result = result + separator + " "
		}
		result = result + string(firstChar) + sentence[1:]
	}

	result = result + "."
	//logs.GetLogger().Info(result)
	return result
}

func FirstLetter2Upper(text string) string {
	text = strings.Trim(text, " ")
	if text == "" {
		return text
	}
	firstChar := byte(unicode.ToUpper(rune(text[0])))
	return string(firstChar) + text[1:]
}

func SearchFloat64FromStr(source string) *float64 {
	re := regexp.MustCompile("[0-9]+.?[0-9]*")
	words := re.FindAllString(source, -1)
	if len(words) > 0 {
		numStr := strings.Trim(words[0], " ")
		result, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil
		}
		return &result
	}

	return nil
}

func ConvertPrice2AttoFil(price string) string {
	fields := strings.Fields(price)
	if len(fields) < 1 {
		return ""
	}
	if len(fields) < 2 {
		return fields[0]
	}
	priceAttoFil, err := decimal.NewFromString(fields[0])
	if err != nil {
		logs.GetLogger().Error()
		return ""
	}
	unit := strings.ToUpper(fields[1])
	switch unit {
	case "FIL":
		priceAttoFil = priceAttoFil.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E18))
	case "MILLIFIL":
		priceAttoFil = priceAttoFil.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E15))
	case "MICROFIL":
		priceAttoFil = priceAttoFil.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E12))
	case "NANOFIL":
		priceAttoFil = priceAttoFil.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E9))
	case "PICOFIL":
		priceAttoFil = priceAttoFil.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E6))
	case "FEMTOFIL":
		priceAttoFil = priceAttoFil.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE_1E3))
	}

	priceAttoFilStr := priceAttoFil.BigInt().String()

	return priceAttoFilStr
}

func GetPriceFormat(price string) string {
	fields := strings.Fields(price)
	if len(fields) < 1 {
		return ""
	}
	if len(fields) < 2 {
		return fields[0]
	}
	priceAttoFil := int64(1)
	unit := strings.ToUpper(fields[1])
	switch unit {
	case "FIL":
		priceAttoFil = priceAttoFil * 1e18
	case "MILLIFIL":
		priceAttoFil = priceAttoFil * 1e15
	case "MICROFIL":
		priceAttoFil = priceAttoFil * 1e12
	case "NANOFIL":
		priceAttoFil = priceAttoFil * 1e9
	case "PICOFIL":
		priceAttoFil = priceAttoFil * 1e6
	case "FEMTOFIL":
		priceAttoFil = priceAttoFil * 1e3
	}

	result := strconv.FormatInt(priceAttoFil, 10)
	result = strings.TrimPrefix(result, "1")

	return result
}

func GetStr(val interface{}) string {
	switch val := val.(type) {
	case float64:
		return strconv.FormatFloat(val, 'e', -1, 64)
	case float32:
		strconv.FormatFloat(float64(val), 'e', -1, 64)
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
	return ""
}

func GetDefaultTaskName() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	randStr := RandStringRunes(letterRunes, 6)
	taskName := "swan-task-" + randStr
	return taskName
}
