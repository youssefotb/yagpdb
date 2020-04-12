package giveaway

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/bot"
	"github.com/jonas747/yagpdb/commands"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/common/scheduledevents2"
	seventsmodels "github.com/jonas747/yagpdb/common/scheduledevents2/models"
	"github.com/jonas747/yagpdb/giveaway/models"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

var _ commands.CommandProvider = (*Plugin)(nil)
var _ bot.BotInitHandler = (*Plugin)(nil)

type GiveawayUpdateEventData struct {
	ID int64 `json:"id"`
}

func (p *Plugin) BotInit() {
	scheduledevents2.RegisterHandler("giveaway_update", GiveawayUpdateEventData{}, p.handleGiveawayUpdateEvent)
}

var emojiRegex = regexp.MustCompile("<:(.*):([0-9]*)>")

func (p *Plugin) AddCommands() {

	categoryGiveaway := &dcmd.Category{
		Name:        "Giveaways",
		Description: "Giveaway commands",
		HelpEmoji:   "🎫",
		EmbedColor:  0x42b9f4,
	}

	// giveaway start
	cmdGiveawayStart := &commands.YAGCommand{
		CmdCategory:  categoryGiveaway,
		Name:         "Start",
		Aliases:      []string{"create", "new"},
		Description:  "Starts a giveaway, duration and title is required.",
		RequiredArgs: 2,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "Duration", Type: &commands.DurationArg{}},
			&dcmd.ArgDef{Name: "Title", Type: dcmd.String, Default: "none"},
		},
		ArgSwitches: []*dcmd.ArgDef{
			&dcmd.ArgDef{Switch: "winners", Name: "Number of Winners", Default: 1, Type: &dcmd.IntArg{Min: 0, Max: 100}},
			&dcmd.ArgDef{Switch: "emoji", Name: "Emoji to be used for the reaction", Default: "🎉", Type: dcmd.String},
			&dcmd.ArgDef{Switch: "color", Name: "Embed color (hex)", Default: "f44242", Type: dcmd.String},
		},

		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			count, err := models.Giveaways(models.GiveawayWhere.GuildID.EQ(parsed.GS.ID), models.GiveawayWhere.EndedAt.IsNotNull()).CountG(parsed.Context())
			if err != nil {
				return nil, err
			}

			if count >= 100 {
				return "Max 100 active giveaways at a time on a server", nil
			}

			emoji := parsed.Switch("emoji").Str()
			colorStr := parsed.Switch("color").Str()
			winners := parsed.Switch("winners").Int()

			colorParsed, _ := strconv.ParseInt(colorStr, 16, 32)

			duration := parsed.Args[0].Value.(time.Duration)
			description := parsed.Args[1].Str()

			m, err := common.BotSession.ChannelMessageSend(parsed.CS.ID, "Creating giveaway....")
			if err != nil {
				return nil, err
			}

			emojiUnicode := ""
			emojiID := int64(0)

			matches := emojiRegex.FindAllStringSubmatch(emoji, 1)
			if len(matches) < 1 {
				emojiUnicode = emoji
			} else {
				emojiUnicode = matches[0][1]
				emojiID, _ = strconv.ParseInt(matches[0][2], 10, 64)
			}

			model := &models.Giveaway{
				MessageID: m.ID,
				GuildID:   parsed.GS.ID,
				ChannelID: parsed.CS.ID,

				Author:      parsed.Msg.Author.ID,
				Description: description,
				Color:       int(colorParsed),
				NumWinners:  winners,

				CreatedAt: time.Now(),
				EndsAt:    time.Now().Add(duration),

				ReactEmojiID:      emojiID,
				ReactEmojiUnicode: emojiUnicode,
			}

			err = model.InsertG(parsed.Context(), boil.Infer())
			if err != nil {
				return nil, errors.Wrap(err, "insertg")
			}

			nextRunTime := CalcNextRunTime(model)
			err = scheduledevents2.ScheduleEvent("giveaway_update", parsed.GS.ID, nextRunTime, &GiveawayUpdateEventData{ID: m.ID})
			if err != nil {
				return nil, err
			}

			err = UpdateGiveawayMessage(model)
			if err != nil {
				return nil, errors.Wrap(err, "updt.msg")
			}

			reactionStr := emojiUnicode
			if emojiID != 0 {
				reactionStr += ":" + strconv.FormatInt(emojiID, 10)
			}

			err = common.BotSession.MessageReactionAdd(parsed.CS.ID, m.ID, reactionStr)
			return nil, errors.Wrap(err, "react")
		},
	}

	// giveaway end
	cmdGiveawayEnd := &commands.YAGCommand{
		CmdCategory:  categoryGiveaway,
		Name:         "End",
		Aliases:      []string{"stop"},
		Description:  "Ends a giveaway, picking a winner straight away",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "MessageID", Type: dcmd.Int},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			giveaway := parsed.Context().Value(CtxKeyGiveaway).(*models.Giveaway)
			if giveaway.EndedAt.Valid {
				return "Giveaway already ended, use `giveaway reroll` if you want to repick the winners.", nil
			}

			if parsed.Msg.Author.ID != giveaway.Author {
				hasPerms, err := bot.AdminOrPermMS(commands.ContextMS(parsed.Context()), parsed.CS.ID, 0)
				if err != nil {
					return nil, err
				}
				if !hasPerms {
					return "You do not have permissions to run this command", nil
				}
			}

			err := CompleteGiveaway(giveaway)
			return nil, err
		},
	}

	// giveaway repick
	cmdGiveawayRepick := &commands.YAGCommand{
		CmdCategory:  categoryGiveaway,
		Name:         "Repick",
		Aliases:      []string{"reroll"},
		Description:  "Repicks winners for this giveaway",
		RequiredArgs: 1,
		Arguments: []*dcmd.ArgDef{
			&dcmd.ArgDef{Name: "MessageID", Type: dcmd.Int},
		},
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			giveaway := parsed.Context().Value(CtxKeyGiveaway).(*models.Giveaway)
			if !giveaway.EndedAt.Valid {
				return "Giveaway has not ended yet, use `giveaway end` if you want to end it.", nil
			}

			if parsed.Msg.Author.ID != giveaway.Author {
				hasPerms, err := bot.AdminOrPermMS(commands.ContextMS(parsed.Context()), parsed.CS.ID, 0)
				if err != nil {
					return nil, err
				}
				if !hasPerms {
					return "You do not have permissions to run this command", nil
				}
			}

			err := CompleteGiveaway(giveaway)
			return nil, err
		},
	}

	// giveaway list
	cmdGiveawayList := &commands.YAGCommand{
		CmdCategory: categoryGiveaway,
		Name:        "List",
		Aliases:     []string{"ls"},
		Description: "Lists active giveaways in your server",
		RunFunc: func(parsed *dcmd.Data) (interface{}, error) {
			activeGiveaways, err := models.Giveaways(
				models.GiveawayWhere.GuildID.EQ(parsed.GS.ID),
				models.GiveawayWhere.EndedAt.IsNull(),
				qm.OrderBy("message_id asc"),
			).AllG(parsed.Context())

			if err != nil {
				return nil, err
			}

			var out strings.Builder
			for _, v := range activeGiveaways {
				when := v.EndsAt.Sub(time.Now())
				out.WriteString(fmt.Sprintf("%s - `%d` - %s\n", v.Description, v.MessageID, common.HumanizeDuration(common.DurationPrecisionSeconds, when)))
			}

			return out.String(), nil
		},
	}

	container := commands.CommandSystem.Root.Sub("giveaway", "giveaways", "g")
	container.NotFound = commands.CommonContainerNotFoundHandler(container, "")

	container.AddCommand(cmdGiveawayStart, cmdGiveawayStart.GetTrigger())
	container.AddCommand(cmdGiveawayEnd, cmdGiveawayEnd.GetTrigger().SetMiddlewares(RequireGiveawayMW))
	container.AddCommand(cmdGiveawayRepick, cmdGiveawayRepick.GetTrigger().SetMiddlewares(RequireGiveawayMW))
	container.AddCommand(cmdGiveawayList, cmdGiveawayList.GetTrigger())
}

type CtxKey int

const CtxKeyGiveaway CtxKey = iota

func RequireGiveawayMW(inner dcmd.RunFunc) dcmd.RunFunc {
	return func(data *dcmd.Data) (interface{}, error) {
		mID := data.Args[0].Int64()

		giveaway, err := models.Giveaways(
			models.GiveawayWhere.GuildID.EQ(data.GS.ID),
			models.GiveawayWhere.MessageID.EQ(mID),
		).OneG(data.Context())

		if err != nil {
			if errors.Cause(err) == sql.ErrNoRows {
				return "Couldn't find any giveaway using that ID", nil
			}

			return "Failed retreiving giveaway", err
		}

		ctx := data.Context()
		ctx = context.WithValue(ctx, CtxKeyGiveaway, giveaway)
		data = data.WithContext(ctx)
		return inner(data)
	}
}

func (p *Plugin) handleGiveawayUpdateEvent(evt *seventsmodels.ScheduledEvent, data interface{}) (retry bool, err error) {
	evtData := data.(*GiveawayUpdateEventData)
	giveaway, err := models.FindGiveawayG(context.Background(), evtData.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}

		return true, err
	}

	if giveaway.EndedAt.Valid {
		// probably used the end command to finish the giveaway
		return false, nil
	}

	if time.Now().After(giveaway.EndsAt) {
		err = CompleteGiveaway(giveaway)
		return scheduledevents2.CheckDiscordErrRetry(err), err
	}

	err = UpdateGiveawayMessage(giveaway)
	if err != nil {
		return scheduledevents2.CheckDiscordErrRetry(err), err
	}

	nextRunTime := CalcNextRunTime(giveaway)
	err = scheduledevents2.ScheduleEvent("giveaway_update", evt.GuildID, nextRunTime, &GiveawayUpdateEventData{ID: evtData.ID})
	if err != nil {
		return true, err
	}

	return false, nil
}

func CompleteGiveaway(giveaway *models.Giveaway) error {
	// fetch all participants
	participants, err := GetAllMessageReactions(giveaway.ChannelID, giveaway.MessageID, giveaway.ReactEmojiUnicode, giveaway.ReactEmojiID)
	if err != nil {
		return err
	}

	winners := PickWinners(participants, giveaway.NumWinners)

	descEmoji := giveaway.ReactEmojiUnicode
	if giveaway.ReactEmojiID != 0 {
		descEmoji = fmt.Sprintf("<:%s:%d>", giveaway.ReactEmojiUnicode, giveaway.ReactEmojiID)
	}

	content := descEmoji + " **GIVEAWAY ENDED** " + descEmoji

	extraMsg := ""
	descStr := ""
	if len(winners) < 1 {
		descStr = "No participants :("
	} else if len(winners) == 1 {
		descStr = "Winner: " + winners[0].Username + "#" + winners[0].Discriminator
		extraMsg = "Congratulations <@" + strconv.FormatInt(winners[0].ID, 10) + "> You won " + giveaway.Description
	} else {
		descStr += "Winners: "
		extraMsg = "Congratulations "
		addedExtraMsg := false

	OUTER:
		for i, v := range winners {
			descStr += "\n`" + v.Username + "#" + v.Discriminator + "`"

			// possibly add to the extra winner message
			for j := 0; j < i; j++ {
				if winners[j].ID == v.ID {
					continue OUTER
				}
			}

			if addedExtraMsg {
				extraMsg += ", "
			}

			addedExtraMsg = true
			extraMsg += "<@" + strconv.FormatInt(v.ID, 10) + ">"
		}

		extraMsg += " you won " + giveaway.Description + "!"
	}

	_, err = common.BotSession.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: giveaway.ChannelID,
		ID:      giveaway.MessageID,
		Content: &content,
		Embed: &discordgo.MessageEmbed{
			Title:       giveaway.Description,
			Description: descStr,
			Color:       giveaway.Color,
			Timestamp:   giveaway.EndsAt.Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Ended",
			},
		},
	})
	if err != nil {
		return err
	}

	if extraMsg != "" {
		_, err = common.BotSession.ChannelMessageSend(giveaway.ChannelID, extraMsg)
	}

	if err != nil {
		return err
	}

	giveaway.EndedAt = null.TimeFrom(time.Now())
	_, err = giveaway.UpdateG(context.Background(), boil.Whitelist("ended_at"))
	return err
}

func PickWinners(participants []*discordgo.User, numWinners int) []*discordgo.User {
	winners := make([]*discordgo.User, 0, numWinners)
	if len(participants) < 1 {
		return winners
	}

	// if there's enough participants to create unique winners, then do so
	uniqueWinners := len(participants) >= numWinners

	for i := 0; i < numWinners; i++ {
		winnerI := rand.Intn(len(participants))
		winner := participants[winnerI]

		winners = append(winners, winner)

		if uniqueWinners {
			participants = append(participants[:winnerI], participants[winnerI+1:]...)
		}
	}

	return winners
}

func GetAllMessageReactions(channelID, messageID int64, emojiUnicode string, emojiID int64) ([]*discordgo.User, error) {
	emojiStr := emojiUnicode
	if emojiID != 0 {
		emojiStr += ":" + strconv.FormatInt(emojiID, 10)
	}

	after := int64(0)

	users := make([]*discordgo.User, 0, 100)

	for {
		reactions, err := common.BotSession.MessageReactions(channelID, messageID, emojiStr, 100, 0, after)
		if err != nil {
			return nil, err
		}

		for _, v := range reactions {
			if v.Bot || v.ID == common.BotUser.ID {
				continue
			}

			users = append(users, v)
		}

		if len(reactions) < 100 {
			break
		} else {
			after = reactions[len(reactions)-1].ID
		}
	}

	return users, nil
}

type CacheKey int

const CacheKeyGiveaways CacheKey = iota

// CacheGetActiveGiveaways retrieves a guilds active giveaways either from the local cache or the database
func CacheGetActiveGiveaways(ctx context.Context, guildID int64) ([]*models.Giveaway, error) {
	gs := bot.State.Guild(true, guildID)
	if gs == nil {
		return nil, bot.ErrGuildNotFound
	}

	v, err := gs.UserCacheFetch(CacheKeyGiveaways, func() (interface{}, error) {
		giveaways, err := models.Giveaways(
			models.GiveawayWhere.GuildID.EQ(guildID),
			models.GiveawayWhere.EndedAt.IsNull()).AllG(ctx)

		return giveaways, err
	})

	if err != nil {
		return nil, err
	}

	return v.(models.GiveawaySlice), nil
}

func UpdateGiveawayMessage(giveaway *models.Giveaway) error {

	descEmoji := giveaway.ReactEmojiUnicode
	if giveaway.ReactEmojiID != 0 {
		descEmoji = fmt.Sprintf("<:%s:%d>", giveaway.ReactEmojiUnicode, giveaway.ReactEmojiID)
	}

	remaining := giveaway.EndsAt.Sub(time.Now())

	content := descEmoji + " **GIVEAWAY** " + descEmoji

	footer := strconv.Itoa(giveaway.NumWinners) + " winner(s) | Ends "

	_, err := common.BotSession.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: giveaway.ChannelID,
		ID:      giveaway.MessageID,
		Content: &content,
		Embed: &discordgo.MessageEmbed{
			Title:       giveaway.Description,
			Description: "React with " + descEmoji + " to enter\nTime remaining: " + common.HumanizeDuration(common.DurationPrecisionSeconds, remaining),
			Color:       giveaway.Color,
			Timestamp:   giveaway.EndsAt.Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: footer,
			},
		},
	})

	return err
}

func CalcNextRunTime(giveaway *models.Giveaway) time.Time {

	durFromNow := giveaway.EndsAt.Sub(time.Now())

	now := time.Now()
	if durFromNow < time.Second*10 {
		return now
	} else if durFromNow < time.Minute {
		return now.Add(time.Second * 2)
	} else if durFromNow < time.Minute*10 {
		return now.Add(time.Minute)
	}

	return now.Add(time.Minute * 10)
}
