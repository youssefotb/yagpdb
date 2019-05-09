package giveaway

//go:generate sqlboiler --no-hooks psql

import (
	"github.com/jonas747/yagpdb/common"
)

type Plugin struct{}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Giveaway",
		SysName:  "giveaway",
		Category: common.PluginCategoryMisc,
	}
}

var logger = common.GetPluginLogger(&Plugin{})

func RegisterPlugin() {
	common.InitSchema(DBSchema, "giveaway")

	common.RegisterPlugin(&Plugin{})
}
