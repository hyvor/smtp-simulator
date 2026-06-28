package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
)

type ActionType int

const (
	ActionTypeSyncResponse ActionType = iota
	ActionTypeSyncResponseAtData
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
	AsyncDelay   time.Duration
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
		AsyncDelay:   250 * time.Millisecond,
	},

	"disabled+async": {
		Type:         ActionTypeAsyncBounce,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 2},
		Message:      "Mailbox disabled",
		AsyncDelay:   250 * time.Millisecond,
	},

	"spam+async": {
		Type:         ActionTypeAsyncBounce,
		Code:         550,
		EnhancedCode: [3]int{5, 7, 1},
		Message:      "Message rejected due to spam content",
		AsyncDelay:   250 * time.Millisecond,
	},

	// At-data bounce responses (250 in RCPT, error in DATA)
	"busy+atdata": {
		Type:         ActionTypeSyncResponseAtData,
		Code:         450,
		EnhancedCode: [3]int{4, 2, 1},
		Message:      "Requested mail action not taken: mailbox busy",
	},

	"tempfail+atdata": {
		Type:         ActionTypeSyncResponseAtData,
		Code:         451,
		EnhancedCode: [3]int{4, 3, 0},
		Message:      "Requested action aborted: local error in processing",
	},

	"missing+atdata": {
		Type:         ActionTypeSyncResponseAtData,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 1},
		Message:      "User unknown",
	},

	"disabled+atdata": {
		Type:         ActionTypeSyncResponseAtData,
		Code:         550,
		EnhancedCode: [3]int{5, 1, 2},
		Message:      "Mailbox disabled",
	},

	"spam+atdata": {
		Type:         ActionTypeSyncResponseAtData,
		Code:         550,
		EnhancedCode: [3]int{5, 7, 1},
		Message:      "Message rejected due to low sender reputation",
	},

	// Complaint response
	"complaint": {
		Type:       ActionTypeAsyncComplaint,
		AsyncDelay: 250 * time.Millisecond,
	},
}

const customPrefix = "custom+"

// customSpec is the JSON shape encoded (hex) in the local part of a
// custom+<hex>@... recipient address.
type customSpec struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	At      string `json:"at"`
}

// parseCustomAction decodes a "custom+<hex>" local part into an Action.
// The message is emitted verbatim (EnhancedCode is NoEnhancedCode), so the
// caller controls the entire response line after the status code, including
// whether an enhanced code appears. On malformed input it returns an
// *smtp.SMTPError (501) so the caller can return it directly on the wire.
func parseCustomAction(local string) (Action, error) {
	encoded := local[len(customPrefix):] // slice from original-case local; hex is case-insensitive
	if encoded == "" {
		return Action{}, customErr("requires a hex-encoded JSON spec: custom+<hex>@domain")
	}

	raw, err := hex.DecodeString(encoded)
	if err != nil {
		return Action{}, customErr("local part is not valid hex")
	}

	var spec customSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return Action{}, customErr("decoded payload is not valid JSON")
	}

	if spec.Status < 200 || spec.Status > 599 {
		return Action{}, customErr("status must be a valid SMTP code (200-599)")
	}

	var actionType ActionType
	switch strings.ToLower(spec.At) {
	case "", "rcpt":
		actionType = ActionTypeSyncResponse
	case "data":
		actionType = ActionTypeSyncResponseAtData
	default:
		return Action{}, customErr(`"at" must be "rcpt" or "data"`)
	}

	return Action{
		Type:         actionType,
		Code:         spec.Status,
		EnhancedCode: EnhancedCode{-1, -1, -1}, // NoEnhancedCode -> emit message verbatim
		Message:      spec.Message,
	}, nil
}

func customErr(reason string) *smtp.SMTPError {
	return &smtp.SMTPError{Code: 501, Message: "Invalid custom action: " + reason}
}
