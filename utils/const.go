package utils

const (
	ADD_WORKER    string = "ADD"
	DELETE_WORKER string = "DELETE"
	UPDATE_WORKER string = "UPDATE"

	DEFAULT_BIND_IP string = "0.0.0.0"

	MSG_STATUS_SUCCESS string = "success"
	MSG_STATUS_ERROR   string = "error"

	ENV_RABBIT_MQ_PWD string = "RABBITMQ_PWD"
	ENV_HAPROXY_PWD   string = "HAPROXY_PWD"

	MESSAGE_TYPE         string = "NETWORK_CFG"
	MESSAGE_DEFAULT_FROM string = "HAProxyUpdater"
	MESSAGE_DEFAULT_TO   string = "workerManager"

	HA_PROXY         string = "haproxy"
	HA_PROXY_UPDATER string = "haproxyUpdater"
)
