package model

import (
	"encoding/json"
	"time"

	"github.com/WALL-EEEEEEE/proxy-service/common"

	"golang.org/x/exp/maps"
)

type PROTO struct {
	common.Enum[string]
	Value string
}

func (proto *PROTO) String() string {
	return proto.Value
}

func (proto *PROTO) Values() []string {
	return maps.Keys(protos)
}

func (proto *PROTO) Index() {
}

func (proto *PROTO) Exists() bool {
	_, ok := protos[proto.Value]
	return ok
}

func (proto *PROTO) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	proto.Value = v
	return nil
}
func (proto PROTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(proto.Value)
}

var (
	PROTO_HTTP      PROTO                  = PROTO{Value: "PROTO_HTTP"}
	PROTO_HTTPS     PROTO                  = PROTO{Value: "PROTO_HTTPS"}
	PROTO_WEBSOCKET PROTO                  = PROTO{Value: "PROTO_WEBSOCKET"}
	PROTO_SOCKET    PROTO                  = PROTO{Value: "PROTO_SOCKET"}
	protos          map[string]interface{} = map[string]interface{}{
		PROTO_HTTP.String():      nil,
		PROTO_HTTPS.String():     nil,
		PROTO_SOCKET.String():    nil,
		PROTO_WEBSOCKET.String(): nil,
	}
)

type STATUS struct {
	common.Enum[string]
	Value string
}

func (status *STATUS) String() string {
	return status.Value
}

func (status *STATUS) Values() []string {
	return maps.Keys(statuses)
}

func (status *STATUS) Index() {
}

func (status *STATUS) Exists() bool {
	_, ok := statuses[status.Value]
	return ok
}

func (status *STATUS) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	status.Value = v
	return nil
}
func (status STATUS) MarshalJSON() ([]byte, error) {
	json_bytes, err := json.Marshal(status.Value)
	return json_bytes, err
}

var (
	STATUS_CREATED     STATUS = STATUS{Value: "STATUS_CREATED"}
	STATUS_CHECKED     STATUS = STATUS{Value: "STATUS_CHECKED"}
	STATUS_UNSPECIFIED STATUS = STATUS{Value: "STATUS_UNSPECIFIED"}

	statuses map[string]interface{} = map[string]interface{}{
		STATUS_CREATED.String(): nil,
		STATUS_CHECKED.String(): nil,
	}
)

type UseConfig struct {
	Psn      string            `json:"psn" validate:"required" example:"psn"`
	Host     string            `json:"host" validate:"required" example:"127.0.0.1"`
	Port     int64             `json:"port,string" validate:"required" example:"80"`
	User     string            `json:"user" validate:"required" example:"user"`
	Password string            `json:"password" validate:"required" example:"password"`
	Extra    map[string]string `json:"extra" validate:"required" example:"{}"`
}

type Attr struct {
	Latency      int64    `json:"latency,string,omitempty"`
	Stability    float64  `json:"stability,omitempty"`
	Availiable   bool     `json:"availiable,omitempty"`
	Anonymous    bool     `json:"anonymous,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Country      string   `json:"country,omitempty"`
	City         string   `json:"city,omitempty"`
	Organization string   `json:"organization,omitempty"`
	Location     string   `json:"location,omitempty"`
	Region       string   `json:"region,omitempty"`
}

type Proxy struct {
	Id         string     `json:"id,omitempty"`
	ProviderId string     `json:"provider_id,omitempty"`
	ApiId      string     `json:"api_id,omitempty"`
	Provider   string     `json:"provider,omitempty"`
	Api        string     `json:"api,omitempty"`
	Proto      []PROTO    `json:"proto,omitempty"`
	Ip         string     `json:"ip,omitempty"`
	Port       int64      `json:"port,omitempty"`
	Ttl        int64      `json:"ttl,omitempty"`
	Status     STATUS     `json:"status,omitempty"`
	Attr       *Attr      `json:"attr,omitempty"`
	UseConfig  *UseConfig `json:"use_config,omitempty"`
	CheckedAt  *time.Time `json:"checked_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	ExpiredAt  *time.Time `json:"expired_at,omitempty"`
}
