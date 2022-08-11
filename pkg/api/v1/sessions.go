package v1

import metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"

type Sessions struct {
	*metav1.ObjectMeta
	SessionKey metav1.TCPIPIdentifier
	PacketsKey string
}
