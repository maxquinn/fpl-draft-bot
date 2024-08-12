package draftbot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/maxquinn/draftbot/completions"
	"golang.org/x/exp/slices"
)

type Element struct {
	NewsAdded                        time.Time `json:"news_added"`
	Added                            time.Time `json:"added"`
	NewsReturn                       any       `json:"news_return"`
	EpNext                           any       `json:"ep_next"`
	DirectFreekicksOrder             any       `json:"direct_freekicks_order"`
	PenaltiesOrder                   any       `json:"penalties_order"`
	SquadNumber                      any       `json:"squad_number"`
	CornersAndIndirectFreekicksOrder any       `json:"corners_and_indirect_freekicks_order"`
	NewsUpdated                      any       `json:"news_updated"`
	ChanceOfPlayingThisRound         any       `json:"chance_of_playing_this_round"`
	PointsPerGameRankType            any       `json:"points_per_game_rank_type"`
	PointsPerGameRank                any       `json:"points_per_game_rank"`
	FormRankType                     any       `json:"form_rank_type"`
	FormRank                         any       `json:"form_rank"`
	EpThis                           any       `json:"ep_this"`
	Threat                           string    `json:"threat"`
	PointsPerGame                    string    `json:"points_per_game"`
	Status                           string    `json:"status"`
	SecondName                       string    `json:"second_name"`
	ExpectedGoals                    string    `json:"expected_goals"`
	ExpectedAssists                  string    `json:"expected_assists"`
	ExpectedGoalInvolvements         string    `json:"expected_goal_involvements"`
	ExpectedGoalsConceded            string    `json:"expected_goals_conceded"`
	WebName                          string    `json:"web_name"`
	Form                             string    `json:"form"`
	Influence                        string    `json:"influence"`
	IctIndex                         string    `json:"ict_index"`
	CornersAndIndirectFreekicksText  string    `json:"corners_and_indirect_freekicks_text"`
	Creativity                       string    `json:"creativity"`
	DirectFreekicksText              string    `json:"direct_freekicks_text"`
	PenaltiesText                    string    `json:"penalties_text"`
	News                             string    `json:"news"`
	FirstName                        string    `json:"first_name"`
	Saves                            int       `json:"saves"`
	CreativityRankType               int       `json:"creativity_rank_type"`
	EventPoints                      int       `json:"event_points"`
	DreamteamCount                   int       `json:"dreamteam_count"`
	DraftRank                        int       `json:"draft_rank"`
	Code                             int       `json:"code"`
	ChanceOfPlayingNextRound         int       `json:"chance_of_playing_next_round"`
	Starts                           int       `json:"starts"`
	YellowCards                      int       `json:"yellow_cards"`
	ID                               int       `json:"id"`
	TotalPoints                      int       `json:"total_points"`
	RedCards                         int       `json:"red_cards"`
	InfluenceRank                    int       `json:"influence_rank"`
	InfluenceRankType                int       `json:"influence_rank_type"`
	CreativityRank                   int       `json:"creativity_rank"`
	Team                             int       `json:"team"`
	ThreatRank                       int       `json:"threat_rank"`
	ThreatRankType                   int       `json:"threat_rank_type"`
	IctIndexRank                     int       `json:"ict_index_rank"`
	IctIndexRankType                 int       `json:"ict_index_rank_type"`
	PenaltiesSaved                   int       `json:"penalties_saved"`
	PenaltiesMissed                  int       `json:"penalties_missed"`
	OwnGoals                         int       `json:"own_goals"`
	Minutes                          int       `json:"minutes"`
	GoalsScored                      int       `json:"goals_scored"`
	GoalsConceded                    int       `json:"goals_conceded"`
	CleanSheets                      int       `json:"clean_sheets"`
	Bps                              int       `json:"bps"`
	Bonus                            int       `json:"bonus"`
	Assists                          int       `json:"assists"`
	ElementType                      int       `json:"element_type"`
	InDreamteam                      bool      `json:"in_dreamteam"`
}

type BootstrapData struct {
	Elements []Element `json:"elements"`
}

type LeagueEntry struct {
	JoinedTime      time.Time `json:"joined_time"`
	EntryName       string    `json:"entry_name"`
	PlayerFirstName string    `json:"player_first_name"`
	PlayerLastName  string    `json:"player_last_name"`
	ShortName       string    `json:"short_name"`
	EntryID         int       `json:"entry_id"`
	ID              int       `json:"id"`
	WaiverPick      int       `json:"waiver_pick"`
}

type LeagueData struct {
	LeagueEntries []LeagueEntry `json:"league_entries"`
}

type PublicTrades struct {
	Trades []Trade `json:"trades"`
}

type Trade struct {
	OfferTime    time.Time `json:"offer_time"`
	ResponseTime time.Time `json:"response_time"`
	State        string    `json:"state"`
	TradeitemSet []struct {
		ElementIn  int `json:"element_in"`
		ElementOut int `json:"element_out"`
	} `json:"tradeitem_set"`
	ID            int `json:"id"`
	OfferedEntry  int `json:"offered_entry"`
	ReceivedEntry int `json:"received_entry"`
}

type PlayerSwap struct {
	In  string
	Out string
}

type Annoucement struct {
	Time     time.Time
	ID       string
	TeamFrom string
	TeamTo   string
	Status   string
	Players  []PlayerSwap
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

func init() {
	godotenv.Load()
	functions.HTTP("LatestTrades", LatestTrades)
}

func colloquialTime(t time.Time) string {
	h := t.Local().Hour()
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
		tradeDeal := completions.TradeDeal{
			Time:             tradeTime,
			Status:           a.Status,
			TeamOffering:     a.TeamFrom,
			TeamReceiving:    a.TeamTo,
			PlayersOffered:   displayIn,
			PlayersRequested: displayOut,
		}
		m, err := completions.CreateCompletion(tradeDeal)
		if err != nil {
			return
		}

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
