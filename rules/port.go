package rules

import (
	"strconv"

	C "github.com/ClashrAuto/Clashr/constant"
)

type Port struct {
	adapter  string
	port     string
	isSource bool
}

func (p *Port) RuleType() C.RuleType {
	if p.isSource {
		return C.SrcPort
	}
	return C.DstPort
}

func (p *Port) IsMatch(metadata *C.Metadata) bool {
	if p.isSource {
		return metadata.SrcPort == p.port
	}
	return metadata.DstPort == p.port
}

func (p *Port) Adapter() string {
	return p.adapter
}

func (p *Port) Payload() string {
	return p.port
}

func NewPort(port string, adapter string, isSource bool) *Port {
	_, err := strconv.Atoi(port)
	if err != nil {
		return nil
	}
	return &Port{
		adapter:  adapter,
		port:     port,
		isSource: isSource,
	}
}
