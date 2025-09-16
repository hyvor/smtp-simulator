package main

import (
	"fmt"
)

type ActionType int

const (
	ActionTypeSyncResponse ActionType = iota
	ActionTypeAsyncBounce
	ActionTypeAsyncComplaint
)

type EnhancedCode [3]int

func (ec EnhancedCode) String() string {
	return fmt.Sprintf("%d.%d.%d", ec[0], ec[1], ec[2])
}

func (ec EnhancedCode) Int() [3]int {
	return [3]int{ec[0], ec[1], ec[2]}
}

type Action struct {
	Type         ActionType
	Code         int
	EnhancedCode EnhancedCode
	Message      string
	AsyncDelay   int // in seconds
}

var localPartToAction = map[string]Action{

	"accept": {
		Type:         ActionTypeSyncResponse,
		Code:         250,
		EnhancedCode: [3]int{2, 0, 0},
		Message:      "OK",
	},

	"busy": {
		Type:         ActionTypeSyncResponse,
		Code:         450,
		EnhancedCode: [3]int{4, 2, 1},
		Message:      "Requested mail action not taken: mailbox busy",
	},

	"tempfail": {
		Type:         ActionTypeSyncResponse,
		Code:         451,
		EnhancedCode: [3]int{4, 3, 0},
		Message:      "Requested action aborted: local error in processing",
	},

	"missing": {
		Type:         ActionTypeSyncResponse,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 1},
		Message:      "User unknown",
	},

	"disabled": {
		Type:         ActionTypeSyncResponse,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 2},
		Message:      "Mailbox disabled",
	},

	"spam": {
		Type:         ActionTypeSyncResponse,
		Code:         550,
		EnhancedCode: [3]int{5, 7, 1},
		Message:      "Message rejected due to low sender reputation",
	},

	// Asynchronous bounce responses
	"missing+async": {
		Type:         ActionTypeAsyncBounce,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 1},
		Message:      "User unknown",
	},

	"disabled+async": {
		Type:         ActionTypeAsyncBounce,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 2},
		Message:      "Mailbox disabled",
	},

	"spam+async": {
		Type:         ActionTypeAsyncBounce,
		Code:         550,
		EnhancedCode: [3]int{5, 7, 1},
		Message:      "Message rejected due to spam content",
	},

	// Complaint response
	"complaint": {
		Type: ActionTypeAsyncComplaint,
	},

	// this is always extended
	"custom": {
		Type: ActionTypeSyncResponse,
	},
}
