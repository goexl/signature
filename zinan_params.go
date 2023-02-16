package songci

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type zinanParams struct {
	*verifierParams

	id        string
	service   string
	product   string
	version   string
	processed headers
	_signed   []string
	payload   []byte
	timestamp int64
}

func newZinanParams(params *verifierParams) *zinanParams {
	return &zinanParams{
		verifierParams: params,

		product:   unknown,
		service:   unknown,
		version:   "1",
		timestamp: time.Now().Unix(),
	}
}

func (zp *zinanParams) secret() (final string) {
	sb := new(strings.Builder)
	sb.WriteString(strings.ToUpper(zp.product))
	sb.WriteString(zp.version)
	sb.WriteString(zp.credential)

	return
}

func (zp *zinanParams) request() string {
	sb := new(strings.Builder)
	sb.WriteString(strings.ToLower(zp.product))
	sb.WriteString(underline)
	sb.WriteString(zp.version)
	sb.WriteString(underline)
	sb.WriteString(request)

	return sb.String()
}

func (zp *zinanParams) unzipRequest(request string) {
	values := strings.Split(request, underline)
	zp.product = values[0]
	zp.version = values[1]
}

func (zp *zinanParams) scope() string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf("%d", zp.timestamp))
	sb.WriteString(slash)
	sb.WriteString(zp.service)
	sb.WriteString(slash)
	sb.WriteString(zp.request())

	return sb.String()
}

func (zp *zinanParams) unzipScope(scope string) (codes []uint8) {
	values := strings.Split(scope, slash)
	if number, pe := strconv.ParseInt(values[0], 10, 64); nil != pe {
		codes = append(codes, codeTimestampFormatError)
	} else {
		zp.timestamp = number
		zp.service = values[1]
		zp.unzipRequest(values[2])
	}

	return
}

func (zp *zinanParams) headers() (headers string) {
	keys := make([]string, 0, len(zp.processed))
	for k := range zp.processed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb := new(strings.Builder)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(equal)
		sb.WriteString(zp.processed[key])
		sb.WriteString(semicolon)
	}

	return sb.String()[:sb.Len()-1]
}

func (zp *zinanParams) signed() (signed string) {
	sort.Strings(zp._signed)
	signed = strings.Join(zp._signed, semicolon)

	return
}

func (zp *zinanParams) unzipSigned(signed string) {
	zp._signed = strings.Split(signed, semicolon)
}

func (zp *zinanParams) validate() (codes []uint8) {
	hasContentType := false
	hasHost := false
	zp.processed = make(headers, len(zp.verifierParams.headers))
	for key, value := range zp.verifierParams.headers {
		newKey := strings.ToLower(key)
		if contentType == newKey {
			hasContentType = true
		}
		if host == newKey {
			hasHost = true
		}
		zp.processed[newKey] = value
	}

	if !hasContentType {
		codes = append(codes, codeNoContentTypeHeader)
	}
	if !hasHost {
		codes = append(codes, codeNoHostHeader)
	}

	return
}
