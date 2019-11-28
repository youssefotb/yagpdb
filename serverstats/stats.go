package serverstats

import "time"

type ChannelStats struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type DailyStats struct {
	ChannelMessages map[string]*ChannelStats `json:"channels_messages"`
	JoinedDay       int                      `json:"joined_day"`
	LeftDay         int                      `json:"left_day"`
	Online          int                      `json:"online_now"`
	TotalMembers    int                      `json:"total_members_now"`
}

func RetrieveDailyStats(guildID int64) (*DailyStats, error) {
	return &DailyStats{}, nil
}

type MemberChartDataPeriod struct {
	T          time.Time `json:"t"`
	Joins      int       `json:"joins"`
	Leaves     int       `json:"leaves"`
	NumMembers int       `json:"num_members"`
	MaxOnline  int       `json:"max_online"`
}

func RetrieveMemberChartStats(guildID int64, days int) ([]*MemberChartDataPeriod, error) {
	return nil, nil
}

type MessageChartDataPeriod struct {
	T            time.Time `json:"t"`
	MessageCount int       `json:"message_count"`
}

func RetrieveMessageChartData(guildID int64, days int) ([]*MessageChartDataPeriod, error) {
	return nil, nil
}
