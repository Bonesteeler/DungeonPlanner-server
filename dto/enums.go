package dto

type ModerationStatus int

const (
    ModerationStatusPending ModerationStatus = iota
    ModerationStatusApproved
    ModerationStatusRejected
)