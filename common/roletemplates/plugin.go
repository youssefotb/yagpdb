package roletemplates

import "github.com/jonas747/yagpdb/common"

type Plugin struct{}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Role Templates",
		SysName:  "roletemplates",
		Category: common.PluginCategoryCore,
	}
}

var logger = common.GetPluginLogger(&Plugin{})

func RegisterPlugin() {
	plugin := &Plugin{}
	common.RegisterPlugin(plugin)
}
