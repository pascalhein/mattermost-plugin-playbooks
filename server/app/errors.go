package app

import "github.com/pkg/errors"

// ErrNotFound used to indicate entity not found.
var ErrNotFound = errors.New("not found")

// ErrChannelDisplayNameInvalid is used to indicate a channel name is too long.
var ErrChannelDisplayNameInvalid = errors.New("channel name is invalid or too long")

// ErrPermission is used to indicate a user does not have permissions
var ErrPermission = errors.New("permissions error")

// ErrPlaybookRunNotActive is used to indicate trying to run a command on an incident that has ended.
var ErrPlaybookRunNotActive = errors.New("incident not active")

// ErrPlaybookRunActive is used to indicate trying to run a command on an incident that is active.
var ErrPlaybookRunActive = errors.New("incident active")

// ErrMalformedPlaybookRun is used to indicate an incident is not valid
var ErrMalformedPlaybookRun = errors.New("incident active")

// ErrDuplicateEntry indicates the db could not make an insert because the entry already existed.
var ErrDuplicateEntry = errors.New("duplicate entry")
