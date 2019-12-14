package roletemplates

import (
	"github.com/jonas747/yagpdb/common"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS role_templates (
	guild_id BIGINT NOT NULL,
	local_id BIGINT NOT NULL,

	builtin BOOL NOT NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL,

	roles BIGINT[],
	
	UNIQUE(guild_id, name),
	PRIMARY KEY(guild_id, local_id)
	);
	`

// GetGuildRoleTemplates returns all role templates for a guild
func GetGuildRoleTemplates(guildID int64) ([]*RoleTemplate, error) {
	const q = `SELECT local_id, builtin, name, description, roles FROM role_templates WHERE guild_id=$1`
	rows, err := common.PQ.Query(q, guildID)
	if err != nil {
		return nil, err
	}

	result := make([]*RoleTemplate, 0)

	for rows.Next() {
		rt := RoleTemplate{
			GuildID: guildID,
		}

		err = rows.Scan(&rt.ID, &rt.Builtin, &rt.Name, &rt.Description, &rt.Roles)
		if err != nil {
			return nil, err
		}

		result = append(result, &rt)
	}

	return result, nil
}

// GetNamedRoleTemplate returns a named role template for the specified guild
func GetNamedRoleTemplate(guildID int64, name string) (*RoleTemplate, error) {
	const q = `SELECT local_id, builtin, name, description, roles FROM role_templates WHERE guild_id=$1 AND name=$2`

	rt := RoleTemplate{
		GuildID: guildID,
	}

	err := common.PQ.QueryRow(q, guildID, name).Scan(&rt.ID, &rt.Builtin, &rt.Name, &rt.Description, &rt.Roles)
	if err != nil {
		return nil, err
	}

	return &rt, nil
}
