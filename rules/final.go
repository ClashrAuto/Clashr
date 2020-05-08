package rules

import (
	C "github.com/ClashrAuto/Clashr/constant"
)

type Match struct {
	adapter string
}

func (f *Match) RuleType() C.RuleType {
	return C.MATCH
}

func (f *Match) IsMatch(metadata *C.Metadata) bool {
	return true
}

func (f *Match) Adapter() string {
	return f.adapter
}

func (f *Match) Payload() string {
	return ""
}

func NewMatch(adapter string) *Match {
	return &Match{
		adapter: adapter,
	}
}
