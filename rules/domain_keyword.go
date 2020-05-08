package rules

import (
	"strings"

	C "github.com/ClashrAuto/Clashr/constant"
)

type DomainKeyword struct {
	keyword string
	adapter string
}

func (dk *DomainKeyword) RuleType() C.RuleType {
	return C.DomainKeyword
}

func (dk *DomainKeyword) IsMatch(metadata *C.Metadata) bool {
	if metadata.AddrType != C.AtypDomainName {
		return false
	}
	domain := metadata.Host
	return strings.Contains(domain, dk.keyword)
}

func (dk *DomainKeyword) Adapter() string {
	return dk.adapter
}

func (dk *DomainKeyword) Payload() string {
	return dk.keyword
}

func NewDomainKeyword(keyword string, adapter string) *DomainKeyword {
	return &DomainKeyword{
		keyword: strings.ToLower(keyword),
		adapter: adapter,
	}
}
