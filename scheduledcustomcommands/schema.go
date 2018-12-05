package scheduledcustomcommands

const DBSchema = `
CREATE TABLE IF NOT EXISTS scheduled_custom_commands (
	id BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,

	guild_id BIGINT NOT NULL,
	channel_id BIGINT NOT NULL,

	-- 1 = every x minute
	-- 2 = every x hour
	interval_type INT NOT NULL,
	interval_value INT NOT NULL,
	interval_at_minute INT NOT NULL,

	interval_excluding_days INT[] NOT NULL,
	interval_excluding_hours INT[] NOT NULL,

	last_run TIMESTAMP WITH TIME ZONE,
	next_run TIMESTAMP WITH TIME ZONE NOT NULL,
	num_ratelimit INT NOT NULL
);

CREATE INDEX IF NOT EXISTS scheduled_custom_commands_guild_idx ON scheduled_custom_commands(guild_id);
`
