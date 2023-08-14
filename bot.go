package draftbot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slices"
)

type Element struct {
	ID                               int       `json:"id"`
	Assists                          int       `json:"assists"`
	Bonus                            int       `json:"bonus"`
	Bps                              int       `json:"bps"`
	CleanSheets                      int       `json:"clean_sheets"`
	Creativity                       string    `json:"creativity"`
	GoalsConceded                    int       `json:"goals_conceded"`
	GoalsScored                      int       `json:"goals_scored"`
	IctIndex                         string    `json:"ict_index"`
	Influence                        string    `json:"influence"`
	Minutes                          int       `json:"minutes"`
	OwnGoals                         int       `json:"own_goals"`
	PenaltiesMissed                  int       `json:"penalties_missed"`
	PenaltiesSaved                   int       `json:"penalties_saved"`
	RedCards                         int       `json:"red_cards"`
	Saves                            int       `json:"saves"`
	Threat                           string    `json:"threat"`
	YellowCards                      int       `json:"yellow_cards"`
	Starts                           int       `json:"starts"`
	ExpectedGoals                    string    `json:"expected_goals"`
	ExpectedAssists                  string    `json:"expected_assists"`
	ExpectedGoalInvolvements         string    `json:"expected_goal_involvements"`
	ExpectedGoalsConceded            string    `json:"expected_goals_conceded"`
	Added                            time.Time `json:"added"`
	ChanceOfPlayingNextRound         int       `json:"chance_of_playing_next_round"`
	ChanceOfPlayingThisRound         any       `json:"chance_of_playing_this_round"`
	Code                             int       `json:"code"`
	DraftRank                        int       `json:"draft_rank"`
	DreamteamCount                   int       `json:"dreamteam_count"`
	EpNext                           any       `json:"ep_next"`
	EpThis                           any       `json:"ep_this"`
	EventPoints                      int       `json:"event_points"`
	FirstName                        string    `json:"first_name"`
	Form                             string    `json:"form"`
	InDreamteam                      bool      `json:"in_dreamteam"`
	News                             string    `json:"news"`
	NewsAdded                        time.Time `json:"news_added"`
	NewsReturn                       any       `json:"news_return"`
	NewsUpdated                      any       `json:"news_updated"`
	PointsPerGame                    string    `json:"points_per_game"`
	SecondName                       string    `json:"second_name"`
	SquadNumber                      any       `json:"squad_number"`
	Status                           string    `json:"status"`
	TotalPoints                      int       `json:"total_points"`
	WebName                          string    `json:"web_name"`
	InfluenceRank                    int       `json:"influence_rank"`
	InfluenceRankType                int       `json:"influence_rank_type"`
	CreativityRank                   int       `json:"creativity_rank"`
	CreativityRankType               int       `json:"creativity_rank_type"`
	ThreatRank                       int       `json:"threat_rank"`
	ThreatRankType                   int       `json:"threat_rank_type"`
	IctIndexRank                     int       `json:"ict_index_rank"`
	IctIndexRankType                 int       `json:"ict_index_rank_type"`
	FormRank                         any       `json:"form_rank"`
	FormRankType                     any       `json:"form_rank_type"`
	PointsPerGameRank                any       `json:"points_per_game_rank"`
	PointsPerGameRankType            any       `json:"points_per_game_rank_type"`
	CornersAndIndirectFreekicksOrder any       `json:"corners_and_indirect_freekicks_order"`
	CornersAndIndirectFreekicksText  string    `json:"corners_and_indirect_freekicks_text"`
	DirectFreekicksOrder             any       `json:"direct_freekicks_order"`
	DirectFreekicksText              string    `json:"direct_freekicks_text"`
	PenaltiesOrder                   any       `json:"penalties_order"`
	PenaltiesText                    string    `json:"penalties_text"`
	ElementType                      int       `json:"element_type"`
	Team                             int       `json:"team"`
}

type BootstrapData struct {
	Elements []Element `json:"elements"`
}

type LeagueEntry struct {
	EntryID         int       `json:"entry_id"`
	EntryName       string    `json:"entry_name"`
	ID              int       `json:"id"`
	JoinedTime      time.Time `json:"joined_time"`
	PlayerFirstName string    `json:"player_first_name"`
	PlayerLastName  string    `json:"player_last_name"`
	ShortName       string    `json:"short_name"`
	WaiverPick      int       `json:"waiver_pick"`
}

type LeagueData struct {
	LeagueEntries []LeagueEntry `json:"league_entries"`
}

type PublicTrades struct {
	Trades []Trade `json:"trades"`
}

type Trade struct {
	ID            int       `json:"id"`
	OfferedEntry  int       `json:"offered_entry"`
	OfferTime     time.Time `json:"offer_time"`
	ReceivedEntry int       `json:"received_entry"`
	ResponseTime  time.Time `json:"response_time"`
	State         string    `json:"state"`
	TradeitemSet  []struct {
		ElementIn  int `json:"element_in"`
		ElementOut int `json:"element_out"`
	} `json:"tradeitem_set"`
}

type PlayerSwap struct {
	In  string
	Out string
}

type Annoucement struct {
	ID       string
	TeamFrom string
	TeamTo   string
	Players  []PlayerSwap
	Time     time.Time
	Status   string
}

var tradeStatuses = map[string]string{
	"o": "Offered",   // Offered - Offer has been made.
	"w": "Withdrawn", // Withdrawn - Proposed trade withdrawn by the instigator.
	"r": "Rejected",  // Rejected - Proposed trade rejected by the recipient.
	"a": "Accepted",  // Accepted - The trade has been accepted but requires approval.
	"i": "Invalid",   // Invalid - A player involved in the proposed trade has been part of another accepted trade.
	"v": "Vetoed",    // Vetoed - The trade was accepted but has been vetoed by the league administrator or managers.
	"e": "Expired",   // Expired - The proposed trade wasn't accepted by the trade deadline.
	"p": "Processed", // Processed - The trade has been made.
}

/*
*

	%[1]s - time
	%[2]s - team A
	%[3]s - team B
	%[4]s - players joining A
	%[5]s - players joining B
*/
var messageTemplates = map[string][]string{
	"Processed": {
		`
@everyone
🚨 Trade Alert 🚨

Deal agreed %[1]s between %[2]s and %[3]s.

We understand it's a swap deal, %[4]s for %[5]s.

Here we go! ✨
---
`,
		`
@everyone
It looks like a done deal! ✅

%[5]s will join %[3]s %[1]s, while %[2]s will welcome %[4]s.

Here we go! ✨
---
`,
		`
@everyone
🚨 BREAKING: %[2]s and %[3]s have officially agreed a deal %[1]s.

%[4]s will be flying to the %[2]s grounds to meet with the owners.

%[5]s will be heading to %[3]s to complete medical evaluation.
---
`,
	},
	"Accepted": {
		`
@everyone
Understood that %[2]s and %[3]s are close to completing a swap deal.

%[4]s for %[5]s. The players have approved the transfer.

It's now up to the league whether or not the deal goes through.

%[2]s and %[3]s hope to get the deal over the line before this week's deadline.

More to come...
---
`,
		`
@everyone
%[3]s have given the green light 🚨🟢

%[5]s will be joining %[3]s %[1]s unless there's a veto from the league.

Understood to be a swap deal. Both parties happy.

%[4]s will fly to %[2]s to complete the deal.
---
`,
		`
@everyone
BREAKING: Deal thought to be nearly done between %[2]s and %[3]s 📈

Only the league has the power to stop this one happening.

%[5]s will be joining %[3]s while %[4]s are going the other way.

%[2]s have already booked medical tests.
---
`,
	},
	"Vetoed": {
		`
@everyone
🛑 Dramatic last minute veto! 🛑

The managers have decided to block the %[4]s for %[5]s swap deal.

It was thought to be a done deal %[1]s between %[2]s and %[3]s.

Both parties are understandably unhappy.
---
`,
		`
@everyone
BREAKING: We thought the saga was over between %[2]s and %[3]s.

The league has put a halt to the swap deal - %[4]s for %[5]s

The %[3]s camp is fuming. %[2]s had done a lot to try and get this one over the line.

It's not to be 📉
---
`,
		`
@everyone
BREAKING %[1]s 🔴 - The %[4]s x %[5]s swap has been vetoed!

The %[2]s agents flew to the %[3]s camp yesterday to try and get the deal over the line.

Understood that the league weren't happy with the deal and have halted it.
---
`,
	},
}

func init() {
	godotenv.Load()
	functions.HTTP("LatestTrades", LatestTrades)
}

func colloquialTime(t time.Time) string {
	h := t.Hour()
	switch {
	case h >= 22 && h <= 3:
		return "late tonight"
	case h >= 4 && h <= 11:
		return "this morning"
	case h >= 13 && h <= 17:
		return "this afternoon"
	case h >= 18 && h <= 21:
		return "this evening"
	default:
		return "today"
	}
}

func fetchGlobalData(b *BootstrapData) error {
	resp, err := http.Get("https://draft.premierleague.com/api/bootstrap-static")
	if err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(b); err != nil {
		return err
	}

	return nil
}

func fetchLeagueData(l *LeagueData) error {
	leagueDataUrl := fmt.Sprintf("https://draft.premierleague.com/api/league/%s/details", os.Getenv("DRAFT_LEAGUE_ID"))
	resp, err := http.Get(leagueDataUrl)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(l); err != nil {
		return err
	}

	return nil
}

func fetchPrevAnnouncements(d *discordgo.Session) []*discordgo.Message {
	p, err := d.ChannelMessages(os.Getenv("META_CHANNEL_ID"), 0, "", "", "")
	if err != nil {
		panic(err)
	}

	return p
}

func postTradeAnnouncements(d *discordgo.Session, a Annoucement) {
	var inPlayers []string
	var outPlayers []string
	var displayIn string
	var displayOut string
	for _, swap := range a.Players {
		inPlayers = append(inPlayers, swap.In)
		outPlayers = append(outPlayers, swap.Out)
	}
	if tradeLen := len(inPlayers); tradeLen > 1 {
		displayIn = strings.Join(inPlayers[:tradeLen-1], ", ") + " and " + inPlayers[tradeLen-1]
		displayOut = strings.Join(outPlayers[:tradeLen-1], ", ") + " and " + outPlayers[tradeLen-1]
	} else {
		displayIn = inPlayers[0]
		displayOut = outPlayers[0]
	}
	tradeTime := colloquialTime(a.Time)
	if a.Status == "Processed" || a.Status == "Accepted" || a.Status == "Vetoed" {
		randIdx := rand.Intn(len(messageTemplates[a.Status]))
		m := fmt.Sprintf(messageTemplates[a.Status][randIdx], tradeTime, a.TeamFrom, a.TeamTo, displayIn, displayOut)
		d.ChannelMessageSend(os.Getenv("ANNOUNCEMENT_CHANNEL_ID"), m)
		d.ChannelMessageSend(os.Getenv("META_CHANNEL_ID"), a.ID+"-"+a.Status)
	}
}

func buildAnnoucment(a *Annoucement, t Trade, b BootstrapData, l LeagueData) {
	offeredIdx := slices.IndexFunc(l.LeagueEntries, func(el LeagueEntry) bool {
		return el.EntryID == t.OfferedEntry
	})
	recievedIdx := slices.IndexFunc(l.LeagueEntries, func(el LeagueEntry) bool {
		return el.EntryID == t.ReceivedEntry
	})
	if offeredIdx != -1 && recievedIdx != -1 {
		a.TeamFrom = l.LeagueEntries[offeredIdx].EntryName
		a.TeamTo = l.LeagueEntries[recievedIdx].EntryName
	}
	for _, swap := range t.TradeitemSet {
		inIdx := slices.IndexFunc(b.Elements, func(el Element) bool {
			return el.ID == swap.ElementIn
		})
		outIdx := slices.IndexFunc(b.Elements, func(el Element) bool {
			return el.ID == swap.ElementOut
		})

		if inIdx != -1 && outIdx != -1 {
			a.Players = append(a.Players, PlayerSwap{
				In:  b.Elements[inIdx].FirstName + " " + b.Elements[inIdx].SecondName,
				Out: b.Elements[outIdx].FirstName + " " + b.Elements[outIdx].SecondName,
			})
		}
	}
	if t.State == "p" || t.State == "v" || t.State == "a" {
		a.Time = t.ResponseTime
	}
	// else offer time?
}

func LatestTrades(w http.ResponseWriter, r *http.Request) {
	bootstrapData := BootstrapData{}
	if err := fetchGlobalData(&bootstrapData); err != nil {
		fmt.Fprint(w, err)
		return
	}

	leagueData := LeagueData{}
	if err := fetchLeagueData(&leagueData); err != nil {
		fmt.Fprint(w, err)
		return
	}

	tradeData := PublicTrades{}
	tradeDataUrl := fmt.Sprintf("https://draft.premierleague.com/api/draft/league/%s/trades", os.Getenv("DRAFT_LEAGUE_ID"))
	resp, err := http.Get(tradeDataUrl)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(&tradeData); err != nil {
		fmt.Fprint(w, err)
		return
	}

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	prevAnnouncements := fetchPrevAnnouncements(discord)
	for _, trade := range tradeData.Trades {
		tradeID := strconv.Itoa(trade.ID)
		status := tradeStatuses[trade.State]
		notPosted := slices.IndexFunc(prevAnnouncements, func(a *discordgo.Message) bool {
			return a.Content == tradeID+"-"+status
		}) == -1

		if notPosted {
			a := Annoucement{
				ID:     tradeID,
				Status: status,
			}
			buildAnnoucment(&a, trade, bootstrapData, leagueData)
			postTradeAnnouncements(discord, a)
		}
	}
}
