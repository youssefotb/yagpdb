package common

import (
	"regexp"
)

type RoleTemplate struct {
	ID          int64
	GuildID     int64
	Name        string
	Description string
	Builtin     bool

	Roles []int64
}

type BuiltinRoleTemplate struct {
	Template *RoleTemplate
	Regex    *regexp.Regexp
}

var (
	builtinRoleTemplates = []BuiltinRoleTemplate{
		BuiltinRoleTemplate{
			Template: &RoleTemplate{
				Name:        "Administrators",
				Description: "The top level of staff on your server",
				Builtin:     true,
			},
			Regex: regexp.MustCompile("(?i)admin(s|istrator)?"),
		},
		BuiltinRoleTemplate{
			Template: &RoleTemplate{
				Name:        "Moderators",
				Description: "Users with some moderating capability",
				Builtin:     true,
			},
			Regex: regexp.MustCompile("(?i)mod(s|erator)?"),
		},
		BuiltinRoleTemplate{
			Template: &RoleTemplate{
				Name:        "Priviledged Users",
				Description: "Users with slightly more capability and trust than normal users",
				Builtin:     true,
			},
		},
	}
)

func GetGuildRoleTemplates() {

}
