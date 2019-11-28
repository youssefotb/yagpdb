package messagestatscollector

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/common"
	"github.com/sirupsen/logrus"
)

// Collector is a message stats collector which will preiodically update the serberstats messages table with stats
type Collector struct {
	MsgEvtChan chan *discordgo.Message

	interval time.Duration

	channels map[int64]*entry
	// buf      []*discordgo.Message
	// channels []int64
	l *logrus.Entry
}

type entry struct {
	GuildID   int64
	ChannelID int64
	Count     int64
}

// NewCollector creates a new Collector
func NewCollector(l *logrus.Entry, updateInterval time.Duration) *Collector {
	col := &Collector{
		MsgEvtChan: make(chan *discordgo.Message, 1000),
		interval:   updateInterval,
		l:          l,
	}

	go col.run()

	return col
}

func (c *Collector) run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case msg := <-c.MsgEvtChan:
			c.handleIncMessage(msg)
		case <-ticker.C:
			err := c.flush()
			if err != nil {
				c.l.Errorf("failed updating temp serverstats: %+v", err)
			}
		}
	}
}

func (c *Collector) handleIncMessage(msg *discordgo.Message) {
	if c, ok := c.channels[msg.ChannelID]; ok {
		c.Count++
		return
	}

	c.channels[msg.ChannelID] = &entry{
		GuildID:   msg.GuildID,
		ChannelID: msg.ChannelID,
		Count:     1,
	}
}

func (c *Collector) flush() error {
	c.l.Debugf("message stats collector is flushing: lc: %d", len(c.channels))
	if len(c.channels) < 1 {
		return nil
	}

	const updateQuery = `
	INSERT INTO server_stats_hourly_periods_messages (guild_id, t, channel_id, count) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (guild_id, channel_id, t) DO UPDATE
	SET count = server_stats_hourly_periods_messages + $4`

	tx, err := common.PQ.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.WithStackIf(err)
	}

	for _, v := range c.channels {

	}

	// 	channelStats := make([]*models.ServerStatsPeriod, 0, len(c.channels))

	// OUTER:
	// 	for _, v := range c.buf {

	// 		for _, cm := range channelStats {
	// 			if cm.ChannelID.Int64 == v.ChannelID {
	// 				cm.Count.Int64++
	// 				continue OUTER
	// 			}
	// 		}

	// 		created, err := v.Timestamp.Parse()
	// 		if err != nil {
	// 			c.l.WithError(err).Errorf("Message has invalid timestamp: %s (%d/%d/%d)", v.Timestamp, v.GuildID, v.ChannelID, v.ID)
	// 			created = time.Now()
	// 		}

	// 		channelModel := &models.ServerStatsPeriod{
	// 			GuildID:   null.Int64From(v.GuildID),
	// 			ChannelID: null.Int64From(v.ChannelID),
	// 			Started:   null.TimeFrom(created), // TODO: we should calculate these from the min max snowflake ids
	// 			Duration:  null.Int64From(int64(time.Minute)),
	// 			Count:     null.Int64From(1),
	// 		}
	// 		channelStats = append(channelStats, channelModel)
	// 	}

	// tx, err := common.PQ.BeginTx(context.Background(), nil)
	// if err != nil {
	// 	return errors.WithStackIf(err)
	// }

	// for _, model := range channelStats {
	// 	err = model.Insert(context.Background(), tx, boil.Infer())
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return errors.WithStackIf(err)
	// 	}
	// }

	// err = tx.Commit()
	// if err != nil {
	// 	return errors.WithStackIf(err)
	// }

	// reset buffers
	c.channels = make(map[int64]*entry)

	return nil
}
