package messagers

/*
Defining queue names as constants which we can refer to when needed. This way we avoid possible issues with typos in the queue name.
Furthermore, we get a centralized way of defining our queues. Clean code 👍👍👍
*/
const (
	EventsQueue = "events"
	LogsQueue   = "logs"
)
