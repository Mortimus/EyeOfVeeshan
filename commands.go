package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/sheets/v4"
)

// BotAction is the function called when a BotCommand is triggered
type BotAction func(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string)

// srv is the global to connect to google sheets
var srv *sheets.Service

// BotCommand contains everything for a bot response to a user
type BotCommand struct {
	command     string    // string to match in discord channel to trigger this command
	help        string    // help text when user asks the bot for help
	action      BotAction // function to call when command is triggered
	dmOnly      bool      // only trigger if called in a DM
	priviledged bool      // requires a priviledged role to activate
	hidden      bool      // do not list in !help
	minParams   int       // Min # of parameters required
}

var botCommands []BotCommand

func init() {
	// Seed that number generator
	rand.Seed(time.Now().UnixNano())

	// TODO: We need to sanitize ALL config changes involving strings (maybe just drop double quotes)
	// TODO: Set minParams for every input
	// TODO: Loading config verify no commands have matching commands
	// TODO: Guild item tracking/giving commands
	// TODO: Config change command is DM only and Priv only

	dkpCommand := BotCommand{
		command:     configuration.CommDKPCommand,
		help:        configuration.CommDKPHelp,
		action:      LookupDKP,
		dmOnly:      configuration.CommDKPDMOnly,
		priviledged: configuration.CommDKPPriv,
		hidden:      configuration.CommDKPHidden,
	}
	botCommands = append(botCommands, dkpCommand)
	//------------------------------------------------
	summaryCommand := BotCommand{
		command:     configuration.CommRaidSummaryCommand,
		help:        configuration.CommRaidSummaryHelp,
		action:      LookupDKPSummary,
		dmOnly:      configuration.CommRaidSummaryDMOnly,
		priviledged: configuration.CommRaidSummaryPriv,
		hidden:      configuration.CommRaidSummaryHidden,
	}
	botCommands = append(botCommands, summaryCommand)
	//------------------------------------------------
	helpCommand := BotCommand{
		command:     configuration.CommHelpCommand,
		help:        configuration.CommHelpHelp,
		action:      Help,
		dmOnly:      configuration.CommHelpDMOnly,
		priviledged: configuration.CommHelpPriv,
		hidden:      configuration.CommHelpHidden,
	}
	botCommands = append(botCommands, helpCommand)
	//------------------------------------------------
	dbrCommand := BotCommand{
		command:     configuration.CommDBRCommand,
		help:        configuration.CommDBRHelp,
		action:      DBR,
		dmOnly:      configuration.CommDBRDMOnly,
		priviledged: configuration.CommDBRPriv,
		hidden:      configuration.CommDBRHidden,
	}
	botCommands = append(botCommands, dbrCommand)
	//------------------------------------------------
	kronoCommand := BotCommand{
		command:     configuration.CommKronoCommand,
		help:        configuration.CommKronoHelp,
		action:      LookupKrono,
		dmOnly:      configuration.CommKronoDMOnly,
		priviledged: configuration.CommKronoPriv,
		hidden:      configuration.CommKronoHidden,
	}
	botCommands = append(botCommands, kronoCommand)
	//------------------------------------------------
	spellCommand := BotCommand{
		command:     configuration.CommSpellCommand,
		help:        configuration.CommSpellHelp,
		action:      GetPlayerSpell,
		dmOnly:      configuration.CommSpellDMOnly,
		priviledged: configuration.CommSpellPriv,
		hidden:      configuration.CommSpellHidden,
	}
	botCommands = append(botCommands, spellCommand)
	//------------------------------------------------
	givespellCommand := BotCommand{
		command:     configuration.CommGiveSpellCommand,
		help:        configuration.CommGiveSpellHelp,
		action:      SetPlayerSpell,
		dmOnly:      configuration.CommGiveSpellDMOnly,
		priviledged: configuration.CommGiveSpellPriv,
		hidden:      configuration.CommGiveSpellHidden,
	}
	botCommands = append(botCommands, givespellCommand)
	//------------------------------------------------
	readRules := BotCommand{
		command:     configuration.CommRulesCommand,
		help:        configuration.CommRulesHelp,
		action:      ReadRules,
		dmOnly:      configuration.CommRulesDMOnly,
		priviledged: configuration.CommRulesPriv,
		hidden:      configuration.CommRulesHidden,
	}
	botCommands = append(botCommands, readRules)
	//------------------------------------------------
	// changeConfig := BotCommand{
	// 	command:     "!config",
	// 	help:        "Let's you modify configuration without modifying code",
	// 	action:      ChangeConfig,
	// 	dmOnly:      true,
	// 	priviledged: true,
	// 	hidden:      true,
	// 	minParams:   2,
	// }
	// botCommands = append(botCommands, changeConfig)
	//------------------------------------------------
	// rollCommand := BotCommand{
	// 	command: "!roll",
	// 	help:    "Roll XdY sided dice : !roll 2d10 to roll 2 ten sided dice.",
	// 	action:  Roll,
	// }
	// botCommands = append(botCommands, rollCommand)
}

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, message []string) string { // prolly replace user with session to check for dm/rank
	l := LogInit("runCommand-commands.go")
	defer l.End()
	if len(message) > 0 && len(message[0]) > 0 && message[0][0] == '!' { // Command attempted
		l.InfoF("Command: %s attempted by %v", message, m.Author)
		for _, command := range botCommands {
			// log.Printf("Command: %s vs message[0]: %s", command.command, message[0])
			if command.command == strings.ToLower(message[0]) { // Command found!
				l.InfoF("Command: %s matches %s", strings.ToLower(message[0]), command.command)
				if command.dmOnly && !ComesFromDM(s, m) {
					l.InfoF("Command is dm only and coming outside of DM's: %s", message)
					return ""
				}
				if command.priviledged && !isPriviledged(s, m.Author.ID) {
					l.WarnF("Command is priviledged only and coming from a nont-privledged user: %s -- %v", message, m.Author)
					return configuration.NoPrivResponse
				}
				response := command.action(s, m, message)
				l.TraceF("Message complete, responding with: %s", response)
				return response
			}
		}
	}
	return ""
}

// Help lists all commands registered
func Help(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("Help-commands.go")
	defer l.End()
	for _, command := range botCommands {
		if !command.hidden {
			response += fmt.Sprintf("%s: %s\n", command.command, command.help)
		} else {
			l.InfoF("Skipping command %s due to being hidden", command.command)
		}
	}
	return response
}

const unRestrictedConfig = 14

// ChangeConfig lets priv modify configuration
// func ChangeConfig(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
// 	// a := &A{Foo: "afoo"}
// 	fmt.Printf("Message: %s\nLen: %d\n", message, len(message))
// 	val := reflect.ValueOf(&configuration).Elem()
// 	if len(message) < 3 {
// 		for i := unRestrictedConfig; i < val.NumField(); i++ {
// 			// fmt.Println(val.Type().Field(i).Name)
// 			response = fmt.Sprintf("%s\n%s :: %v", response, val.Type().Field(i).Name, reflect.ValueOf(configuration).FieldByName(val.Type().Field(i).Name))
// 		}
// 	}
// 	if len(message) == 2 {
// 		for i := unRestrictedConfig; i < val.NumField(); i++ {
// 			// fmt.Println(val.Type().Field(i).Name)
// 			if val.Type().Field(i).Name == message[1] {
// 				response = fmt.Sprintf("%s :: %v (CanSet: %t)", val.Type().Field(i).Name, reflect.ValueOf(&configuration).FieldByName(val.Type().Field(i).Name), reflect.ValueOf(configuration).FieldByName(val.Type().Field(i).Name).CanSet())
// 			}
// 		}
// 	}
// 	if len(message) > 2 {
// 		changeTo := strings.Join(message[2:], " ")
// 		// fmt.Printf("Config to Change: %s\nValue to change to: %s\n", message[1], changeTo)
// 		response = fmt.Sprintf("Setting %s to %s", message[1], changeTo)
// 		// reflect.ValueOf(configuration).FieldByName(val.Type().Field(i).Name) = changeTo
// 	}

// 	return response
// }

// // Roll provides x random numbers from 1-y
// func Roll(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
// 	diceS := strings.Split(message[1], "d")
// 	dice, err := strconv.Atoi(diceS[0])
// 	sides, err := strconv.Atoi(diceS[1])
// 	if err != nil {
// 		response = fmt.Sprintf("Sorry %s i couldn't figure out how to roll those dice.", user.Username)
// 	} else {
// 		if dice > configuration.maxDice {
// 			response = fmt.Sprintf("Please roll less dice %s.", user.Username)
// 			return response
// 		}
// 		if sides < 1 {
// 			response = fmt.Sprintf("Please roll bigger dice %s.", user.Username)
// 			return response
// 		}
// 		for i := 1; i < dice+1; i++ {
// 			response += fmt.Sprintf("Die #%d rolled %d\n", i, rand.Intn(sides)+1)
// 		}
// 	}
// 	return response
// }

// LookupKrono reaches out to araduneauctions to find the 3 day value of krono
func LookupKrono(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("LookupKrono-commands.go")
	defer l.End()
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get("https://api.araduneauctions.net/GetKronoPrice")
	if err != nil {
		// return err
		l.ErrorF("Error getting Kronos: %s", err.Error())
		return "Error connecting to araduneauctions.net"
	}
	defer r.Body.Close()

	val, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.ErrorF("Error getting Krono from response: %s\n", err.Error())
		return "Error obtaining Krono value, please try again later"
	}
	response = "The average price of krono is " + string(val) + " pp"
	return response
}

// DBR Reminds us who is dark blue
func DBR(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("DBR-commands.go")
	defer l.End()
	return "Sinidan is the Dark Blue Rogue"
}

// Player is an entry on the DKP sheet
type Player struct {
	class      string
	rank       string
	name       string
	level      string
	lastRaid   string
	attendance string
	dkp        string
}

// LookupDKP find the message[1] user's DKP on the known google spreadsheet
func LookupDKP(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("LookupDKP-commands.go")
	defer l.End()
	if len(message) > 1 {
		result := lookupPlayer(message[1])
		return result.dkp
	} else {
		l.ErrorF("DKP command ran without a player: %s", message)
	}
	return ""
}

func lookupPlayer(tar string) Player {
	l := LogInit("lookupPlayer-commands.go")
	defer l.End()
	tar = strings.ToLower(tar)
	tar = strings.Title(tar) // Capitilize first letter
	// player := &Player{}
	var player Player
	player.attendance = "No Attendance Found"
	player.class = "Unknown"
	player.dkp = "0"
	player.lastRaid = "No Raids"
	player.level = "0"
	player.rank = "Unknown"
	spreadsheetID := configuration.DKPSheetURL
	readRange := configuration.DKPSRosterSheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		l.ErrorF("Unable to retrieve data from sheet: %v", err)
		return Player{}
		// log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		l.ErrorF("No player lookup response: %v", resp)
		// log.Println("No data found.")
	} else {
		// var lastClass string
		for _, row := range resp.Values {
			// if row[0] == "Necromancer" {
			// 	fmt.Printf("%s: %s\n", row[2], row[6])
			// }
			// l.TraceF("Player: %s Target: %s", row[configuration.DKPSRosterSheetPlayerCol], strings.TrimSpace(tar))
			name := fmt.Sprintf("%s", row[configuration.DKPSRosterSheetPlayerCol])
			if strings.TrimSpace(name) == strings.TrimSpace(tar) {
				player.class = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetClassCol])
				player.rank = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetRankCol])
				player.name = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetPlayerCol])
				player.level = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetLevelCol])
				// player = Player{
				// 	class:      fmt.Sprintf("%v", row[configuration.DKPSRosterSheetClassCol]),
				// 	rank:       fmt.Sprintf("%v", row[configuration.DKPSRosterSheetRankCol]),
				// 	name:       fmt.Sprintf("%v", row[configuration.DKPSRosterSheetPlayerCol]),
				// 	level:      fmt.Sprintf("%v", row[configuration.DKPSRosterSheetLevelCol]),
				// 	// lastRaid:   fmt.Sprintf("%v", row[configuration.DKPSheetLastRaidCol]),
				// 	// attendance: fmt.Sprintf("%v", row[configuration.DKPSheetAttendanceCol]),
				// 	// dkp:        fmt.Sprintf("%v", row[configuration.DKPSheetDKPCol]),
				// }
				break
			}
		}
		if player.name == "" {
			l.ErrorF("Player not found on roster - %s", tar)
		}
	}
	readRange = configuration.DKPSheetName
	resp2, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		l.ErrorF("Unable to retrieve data from sheet 2nd pass: %v", err)
		return Player{}
		// log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp2.Values) == 0 {
		l.ErrorF("No player lookup response: %v", resp)
		// log.Println("No data found.")
	} else {
		// var lastClass string
		for _, row := range resp2.Values {
			// if row[0] == "Necromancer" {
			// 	fmt.Printf("%s: %s\n", row[2], row[6])
			// }
			name := fmt.Sprintf("%s", row[configuration.DKPSheetNameCol])
			if name == strings.TrimSpace(tar) {
				// player.class = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetClassCol])
				// player.rank = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetRankCol])
				// player.name = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetPlayerCol])
				// player.level = fmt.Sprintf("%v", row[configuration.DKPSRosterSheetLevelCol])

				player.lastRaid = fmt.Sprintf("%v", row[configuration.DKPSheetLastRaidCol])
				player.attendance = fmt.Sprintf("%v", row[configuration.DKPSheetAttendanceCol])
				player.dkp = fmt.Sprintf("%v", row[configuration.DKPSheetDKPCol])
				// player = Player{
				// 	class:      fmt.Sprintf("%v", row[configuration.DKPSRosterSheetClassCol]),
				// 	rank:       fmt.Sprintf("%v", row[configuration.DKPSRosterSheetRankCol]),
				// 	name:       fmt.Sprintf("%v", row[configuration.DKPSRosterSheetPlayerCol]),
				// 	level:      fmt.Sprintf("%v", row[configuration.DKPSRosterSheetLevelCol]),
				// 	// lastRaid:   fmt.Sprintf("%v", row[configuration.DKPSheetLastRaidCol]),
				// 	// attendance: fmt.Sprintf("%v", row[configuration.DKPSheetAttendanceCol]),
				// 	// dkp:        fmt.Sprintf("%v", row[configuration.DKPSheetDKPCol]),
				// }
				return player
			}
		}
		l.ErrorF("Player not found on DKP listing - %s", tar)
	}

	return player
}

// LookupDKPSummary returns a raids summary for a specific player
func LookupDKPSummary(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("LookupDKPSummary-commands.go")
	defer l.End()
	if len(message) > 2 {
		player := message[1]
		player = strings.ToLower(player)
		player = strings.Title(player) // Capitilize first letter
		raid := message[2]
		spreadsheetID := configuration.DKPSheetURL
		readRange := configuration.DKPSummarySheetName
		resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
		if err != nil {
			l.ErrorF("Unable to retrieve data from sheet: %v", err)
			return ""
			// log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}
		response = fmt.Sprintf("%s on %s\n", player, raid)

		if len(resp.Values) == 0 {
			l.ErrorF("No data found. %v", resp)
		} else {
			found := false
			var foundrow int
			for i, row := range resp.Values {
				// if row[0] == "Necromancer" {
				// 	fmt.Printf("%s: %s\n", row[2], row[6])
				// }
				if row[configuration.DKPSummarySheetPlayerCol] == strings.TrimSpace(player) && strings.Contains(fmt.Sprintf("%s", row[configuration.DKPSummarySheetDateCol]), raid) {
					// fmt.Printf("Found row! :: %+v\n", row)
					found = true
					response = fmt.Sprintf("%s%s :: %s\n", response, row[configuration.DKPSummarySheetDKPDescCol], row[configuration.DKPSummarySheetDKPCol])
					foundrow = i
				} else {
					// fmt.Printf("Found not row!\n")
					// for d, vals := range row {
					// 	fmt.Printf("%d: %s\n", d, vals)
					// }
					if found && foundrow == i-1 {
						response = fmt.Sprintf("%s\nTotal :: %s\n", response, row[configuration.DKPSummarySheetDKPCol])
						found = false
						return response
					}
				}
			}
		}
		return response
	}
	return ""
}

func lookupPlayerSpell(player, class, spell string) (bool, string, error) { // Todo: Return Spell Name
	l := LogInit("lookupPlayerSpell-commands.go")
	defer l.End()
	if player == "" || class == "" || spell == "" {
		return false, "", errors.New("Player, class, or spell is nil")
	} else {
		log.Printf("Player: %s Class: %s Spell: %s", player, class, spell)
	}
	spreadsheetID := configuration.SpellSheet
	readRange := class // change based on class TODO: Check against known class names
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		l.ErrorF("Unable to retrieve data from sheet: %v", err)
		return false, "Unable to retrieve data at this time", errors.New("Unable to retrieve data from sheet")
		// log.Fatalf("Unable to retrieve data from sheet: %v", err) // TODO: Gracefully fail
	}

	if len(resp.Values) == 0 {
		l.ErrorF("No data found in response: %v", resp)
	} else {
		var playerColumn int
		for i, row := range resp.Values {
			if len(row) < 1 {
				// return false, "", errors.New("Spell not found")
				continue
			}
			if i == configuration.SpellSheetHeaderRow { // header row, find player
				for col, match := range row {
					if match == player {
						playerColumn = col
					}
				}
				if playerColumn == 0 { // never changed, we can stop searching we didn't find them
					return false, "", errors.New("Player not found")
				}
			}
			spellName := fmt.Sprintf("%v", row[configuration.SpellSheetSpellCol])
			if i > 2 && strings.Contains(spellName, spell) { // if the row containing spell names matches the searched for spell // TODO: We need to do an exact match cause mage spells are DUMB
				col, err := ColumnNumberToName(playerColumn + 1)
				if err != nil {
					return false, "", errors.New("Error converting column number to name")
				}
				tarCell := fmt.Sprintf("%s%d", col, i+1)
				if row[playerColumn] == "TRUE" {
					return true, tarCell, nil
				}
				return false, tarCell, nil
			}
		}
		return false, "", errors.New("Spell not found")
	}
	return false, "", errors.New("Could not pull data from the spreadsheet")
}

// GetPlayerSpell returns if a player already has a spell
func GetPlayerSpell(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("GetPlayerSpell-commands.go")
	defer l.End()
	if len(message) > 2 {
		player := lookupPlayer(message[1])
		l.InfoF("Player: %s = %+v", message[1], player)
		// spellString := message[1:]
		spellString := strings.Join(message[2:], " ")
		hasSpell, _, err := lookupPlayerSpell(message[1], player.class, spellString)
		if err != nil {
			l.ErrorF("Error lookup up spell: %s\n", err.Error())
			return ""
		}
		if hasSpell {
			response = fmt.Sprintf("%s has %s", message[1], spellString)
		} else {
			response = fmt.Sprintf("%s does not have %s", message[1], spellString)
		}

		return response
	}
	return ""
}

// ColumnNumberToName converts from a sane number to an excel letter combination
func ColumnNumberToName(num int) (string, error) {
	l := LogInit("ColumnNumberToName-commands.go")
	defer l.End()
	if num < 1 {
		return "", fmt.Errorf("incorrect column number %d", num)
	}
	// if num > TotalColumns {
	// 	return "", fmt.Errorf("column number exceeds maximum limit")
	// }
	var col string
	for num > 0 {
		col = string(rune((num-1)%26+65)) + col
		num = (num - 1) / 26
	}
	return col, nil
}

func writeToSheet(sheet, cell, value string) {
	l := LogInit("writeToSheet-commands.go")
	defer l.End()
	var vr sheets.ValueRange

	myval := []interface{}{value}
	vr.Values = append(vr.Values, myval)

	_, err := srv.Spreadsheets.Values.Update(sheet, cell, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		l.ErrorF("Unable to retrieve data from sheet. %v", err)
	}
}

// SetPlayerSpell updates the spell spreadsheet
func SetPlayerSpell(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("SetPlayerSpell-commands.go")
	defer l.End()
	if len(message) > 2 {
		player := lookupPlayer(message[1])
		l.InfoF("Player: %s = %+v", message[1], player)
		// spellString := message[1:]
		spellString := strings.Join(message[2:], " ")
		_, cell, err := lookupPlayerSpell(message[1], player.class, spellString)
		if err != nil {
			l.ErrorF("Error lookup up spell: %s\n", err.Error())
			return ""
		}
		// has := fmt.Sprintf("%t", hasSpell)
		writeToSheet(configuration.SpellSheet, player.class+"!"+cell, "TRUE")
		l.InfoF("%s has been given %s by %v", message[1], spellString, m.Author)
		response = fmt.Sprintf("%s has been given %s", message[1], spellString) // TODO: Log this for shaming and blaming
		return response
	}
	return ""
}

// ReadRules pulls the rules from the spreadsheet for player reading
func ReadRules(s *discordgo.Session, m *discordgo.MessageCreate, message []string) (response string) {
	l := LogInit("SetPlayerSpell-commands.go")
	defer l.End()
	spreadsheetID := configuration.SpellSheet
	readRange := configuration.RulesSheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		l.ErrorF("Unable to retrieve data from sheet: %v", err)
		return "Unable to read rules at this time"
		// log.Printf("Unable to retrieve data from sheet: %v", err)
	}
	// log.Printf("User reading rules\n%+v", user)
	if len(resp.Values) == 0 {
		l.ErrorF("No data in sheet response: %v", resp)
	} else {
		// log.Printf("Rules found, reading")
		for _, row := range resp.Values {
			// if row[0] == "Necromancer" {
			// 	fmt.Printf("%s: %s\n", row[2], row[6])
			// }
			for _, part := range row {
				response = fmt.Sprintf("%s\n%v", response, part)
			}
			// TODO: Let player search by rule header
			// log.Printf("Row: %s\n", row)
		}
		// log.Printf("Rules read, printing")
		return response
	}
	return "Unable to read from Google Sheets"
}
