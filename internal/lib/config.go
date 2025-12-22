package lib

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
)

type EtcdVersion string

const (
	ETCD_V2 EtcdVersion = "v2"
	ETCD_V3 EtcdVersion = "v3"
)
