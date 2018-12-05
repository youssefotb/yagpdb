package scheduledcustomcommands

import (
	"github.com/jonas747/yagpdb/common"
	"github.com/sirupsen/logrus"
)

//go:generate sqlboiler --no-hooks psql

type Plugin struct{}

func RegisterPlugin() {
	_, err := common.PQ.Exec(DBSchema)
	if err != nil {
		logrus.WithError(err).Error("failed initializing database schema for scheduledccs, will not be enabled")
		return
	}

	common.RegisterPlugin(&Plugin{})
}

func (p *Plugin) Name() string {
	return "scheduledcustomcommands"
}
