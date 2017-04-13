package hello

type Base struct {
	Status            string   `json:"status"`
	ServerProcessTime string   `json:"process_time"`
	ErrorMessage      []string `json:"error,omitempty"`
	StatusMessage     []string `json:"status,omitempty"`
}

type Response struct {
	Base
	Data interface{} `json:"data"`
}
