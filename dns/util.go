package dns

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"time"

	"github.com/ClashrAuto/Clashr/common/cache"
	"github.com/ClashrAuto/Clashr/log"
	yaml "gopkg.in/yaml.v2"

	D "github.com/miekg/dns"
)

var (
	// EnhancedModeMapping is a mapping for EnhancedMode enum
	EnhancedModeMapping = map[string]EnhancedMode{
		NORMAL.String():  NORMAL,
		FAKEIP.String():  FAKEIP,
		MAPPING.String(): MAPPING,
	}
)

const (
	NORMAL EnhancedMode = iota
	FAKEIP
	MAPPING
)

type EnhancedMode int

// UnmarshalYAML unserialize EnhancedMode with yaml
func (e *EnhancedMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tp string
	if err := unmarshal(&tp); err != nil {
		return err
	}
	mode, exist := EnhancedModeMapping[tp]
	if !exist {
		return errors.New("invalid mode")
	}
	*e = mode
	return nil
}

// MarshalYAML serialize EnhancedMode with yaml
func (e EnhancedMode) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(e.String())
}

// UnmarshalJSON unserialize EnhancedMode with json
func (e *EnhancedMode) UnmarshalJSON(data []byte) error {
	var tp string
	json.Unmarshal(data, &tp)
	mode, exist := EnhancedModeMapping[tp]
	if !exist {
		return errors.New("invalid mode")
	}
	*e = mode
	return nil
}

// MarshalJSON serialize EnhancedMode with json
func (e EnhancedMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e EnhancedMode) String() string {
	switch e {
	case NORMAL:
		return "normal"
	case FAKEIP:
		return "fake-ip"
	case MAPPING:
		return "redir-host"
	default:
		return "unknown"
	}
}

func putMsgToCache(c *cache.Cache, key string, msg *D.Msg) {
	var ttl time.Duration
	if len(msg.Answer) != 0 {
		ttl = time.Duration(msg.Answer[0].Header().Ttl) * time.Second
	} else if len(msg.Ns) != 0 {
		ttl = time.Duration(msg.Ns[0].Header().Ttl) * time.Second
	} else if len(msg.Extra) != 0 {
		ttl = time.Duration(msg.Extra[0].Header().Ttl) * time.Second
	} else {
		log.Debugln("[DNS] response msg error: %#v", msg)
		return
	}

	c.Put(key, msg.Copy(), ttl)
}

func setMsgTTL(msg *D.Msg, ttl uint32) {
	for _, answer := range msg.Answer {
		answer.Header().Ttl = ttl
	}

	for _, ns := range msg.Ns {
		ns.Header().Ttl = ttl
	}

	for _, extra := range msg.Extra {
		extra.Header().Ttl = ttl
	}
}

func isIPRequest(q D.Question) bool {
	if q.Qclass == D.ClassINET && (q.Qtype == D.TypeA || q.Qtype == D.TypeAAAA) {
		return true
	}
	return false
}

func transform(servers []NameServer) []resolver {
	ret := []resolver{}
	for _, s := range servers {
		if s.Net == "https" {
			ret = append(ret, &dohClient{url: s.Addr})
			continue
		}

		ret = append(ret, &client{
			Client: &D.Client{
				Net: s.Net,
				TLSConfig: &tls.Config{
					ClientSessionCache: globalSessionCache,
					// alpn identifier, see https://tools.ietf.org/html/draft-hoffman-dprive-dns-tls-alpn-00#page-6
					NextProtos: []string{"dns"},
				},
				UDPSize: 4096,
			},
			Address: s.Addr,
		})
	}
	return ret
}
