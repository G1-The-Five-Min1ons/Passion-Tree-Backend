package model

type NotificationRecipient struct {
	UserID    string
	Email     string
	FirstName string
}

type WeeklyProgressRow struct {
	NotificationRecipient
	CompletedNodes int
	ActiveDays     int
}

type CommentNotificationRow struct {
	NotificationRecipient
	NewComments int
}

type PlatformAnnouncement struct {
	Title   string
	Content string
}
