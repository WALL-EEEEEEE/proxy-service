package event

type Event = string

const (
	EVENT_CHECKAPI_CREATED Event = "check_api_created"
	EVENT_CHECKAPI_UPDATED Event = "check_api_updated"
	EVENT_CHECKAPI_DELETED Event = "check_api_deleted"
	EVENT_PROXY_CREATED    Event = "proxy_created"
	EVENT_PROXY_DELETED    Event = "proxy_deleted"
	EVENT_PROXY_UPDATED    Event = "proxy_updated"
)
