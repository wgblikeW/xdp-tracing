package v1

import (
	metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
)

type Session struct {
	metav1.TCPIPIdentifier `bson:",inline"`
	InstanceID             string `bson:"identifier"`
	Ttl                    int32  `bson:"ttl"`
	TcpFlagS               string `bson:"tcp_flags"`
	PayloadExist           bool   `bson:"payload_exist"`
	PayloadLen             int32  `bson:"payload_len"`
	Payload                string `bson:"payload"`
	Timestamp              int64  `bson:"timestamp"`
	Direction              string `bson:"direction"`
}
