package structures

type Request struct {
	KeyFamily string
	Key       string
	Value     string
	Action    string
}

type Response struct {
	StatusOk bool
	Message  []string
}
