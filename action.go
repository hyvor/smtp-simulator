package main

type ActionType int

const (
	ActionTypeSyncResponse ActionType = iota
	ActionTypeAsyncBounce
	ActionTypeAsyncComplaint
)

type Action struct {
	Type         ActionType
	StatusCode   int
	EnhancedCode [3]int
	Message      string
	AsyncDelay   int // in seconds
}

var localPartToAction = map[string]Action{

	"accept": {
		Type:         ActionTypeSyncResponse,
		StatusCode:   250,
		EnhancedCode: [3]int{2, 0, 0},
		Message:      "OK",
	},
}
