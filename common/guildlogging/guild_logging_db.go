package serverlogging

import (
	"database/sql"
	"emperror.dev/errors"
	"time"
)

var DBSchemas = []string{
	`
CREATE TABLE IF NOT EXISTS guild_logs (
	id BIGSERIAL PRIMARY KEY,
	guild_id BIGINT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE NOT NULL,

	user_id BIGINT NOT NULL,
	channel_id BIGINT NOT NULL,
	type SMALLINT NOT NULL,
	action TEXT NOT NULL,
)
	`, `
CREATE INDEX IF NOT EXISTS guild_logs_volatile_created_at_idx ON guild_logs(created_at)
	WHERE type != 0; -- An index on all volatile log entries
	`,
}

type GuildLogEntry struct {
	ID        int64
	GuildID   int64
	CreatedAt time.Time

	Plugin    string
	UserID    int64
	ChannelID int64
	Type      LogType
	Action    string
}

type LogType int16

const (
	LogTypeCPAction LogType = iota
	LogTypeError
	LogTypeCommand
	LogTypeOther
)

func ScanGuildEntries(rows *sql.Rows) ([]*GuildLogEntry, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	var result []*GuildLogEntry
	for rows.Next() {
		entry := &GuildLogEntry{}
		scanSlice := make([]interface{}, len(columns))
		for i, v := range columns {
			switch v {
			case "id":
				scanSlice[i] = &entry.ID
			case "guild_id":
				scanSlice[i] = &entry.GuildID
			case "created_at":
				scanSlice[i] = &entry.CreatedAt
			case "user_id":
				scanSlice[i] = &entry.UserID
			case "channel_id":
				scanSlice[i] = &entry.ChannelID
			case "type":
				scanSlice[i] = &entry.Type
			case "action":
				scanSlice[i] = &entry.Action
			}
		}

		err = rows.Scan(scanSlice...)
		if err != nil {
			return nil, errors.WithStackIf(err)
		}
	}

	return result, nil
}
