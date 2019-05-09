package giveaway

const DBSchema = `
CREATE TABLE IF NOT EXISTS giveaways  (
	message_id BIGINT PRIMARY KEY,
	guild_id BIGINT NOT NULL,
	channel_id BIGINT NOT NULL,

	author BIGINT NOT NULL,
	description TEXT NOT NULL,
	num_winners INT NOT NULL,
		
	react_emoji_id BIGINT NOT NULL,
	react_emoji_unicode TEXT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	ends_at TIMESTAMP WITH TIME ZONE NOT NULL,

	ended_at TIMESTAMP WITH TIME ZONE,
	winners BIGINT[],

	color INT NOT NULL
);
`
