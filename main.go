package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/sheets/v4"
)

// srv is the global to connect to google sheets
var srv *sheets.Service

// cal is the global to connect to google calendar
var cal *calendar.Service

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	l := LogInit("getClient-main.go")
	defer l.End()
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	// tokFile := "token.json"
	l.InfoF("Fake loading token from file")
	tok, err := tokenFromFile("")
	if err != nil {
		l.InfoF("Token failed to load, loading from web")
		tok = getTokenFromWeb(config)
		l.InfoF("Saving token")
		saveToken("", tok)
	}
	l.DebugF("Using Token: %+v", tok)
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	l := LogInit("getTokenFromWeb-main.go")
	defer l.End()
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	l.InfoF("Requesting user navigate to: %s", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		l.FatalF("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		l.FatalF("Unable to retrieve token from web: %v", err)
	}
	l.InfoF("Return token: %+v", tok)
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	l := LogInit("tokenFromFile-main.go")
	defer l.End()
	// f, err := os.Open(file)
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	tok := &oauth2.Token{}
	tok.AccessToken = configuration.AccessToken
	tok.Expiry = configuration.Expiry
	tok.RefreshToken = configuration.RefreshToken
	tok.TokenType = configuration.TokenType
	// err = json.NewDecoder(f).Decode(tok)
	l.InfoF("Returning token: %+v", tok)
	return tok, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	l := LogInit("saveToken-main.go")
	defer l.End()
	// fmt.Printf("Saving credential file to: %s\n", path)
	// f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	// if err != nil {
	// 	log.Fatalf("Unable to cache oauth token: %v", err)
	// }
	// defer f.Close()
	// json.NewEncoder(f).Encode(token)
	configuration.AccessToken = token.AccessToken
	configuration.Expiry = token.Expiry
	configuration.RefreshToken = token.RefreshToken
	configuration.TokenType = token.TokenType
	l.InfoF("Saved token to configuration")
	saveConfig()
}

// Inst is an installed struct for google
type Inst struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderx509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
}

// Gtoken is required by google
type Gtoken struct {
	Installed Inst `json:"installed"`
}

func main() {
	// Open Configuration and set log output
	configFile, err := os.OpenFile(configuration.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer configFile.Close()
	log.SetOutput(configFile)
	l := LogInit("main-main.go")
	defer l.End()
	gtoken := &Gtoken{
		Installed: Inst{
			ClientID:                configuration.ClientID,
			ProjectID:               configuration.ProjectID,
			AuthURI:                 configuration.AuthURI,
			TokenURI:                configuration.TokenURI,
			AuthProviderx509CertURL: configuration.AuthProviderx509CertURL,
			ClientSecret:            configuration.ClientSecret,
			RedirectURIs:            configuration.RedirectURIs,
		},
	}
	l.InfoF("Marshalling gToken: %+v", gtoken)
	bToken, err := json.Marshal(gtoken)
	if err != nil {
		l.FatalF("error marshalling gtoken")
	}

	// b, err := ioutil.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(bToken, sheets.SpreadsheetsScope, calendar.CalendarReadonlyScope)
	if err != nil {
		l.FatalF("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err = sheets.New(client)
	if err != nil {
		l.FatalF("Unable retrieve Sheets client: %v", err)
	}

	cal, err = calendar.New(client)
	if err != nil {
		l.FatalF("Unable retrieve Calendar client: %v", err)
	}

	// Create a new Discord session using the provided bot token. os.Getenv("DiscordToken")
	dg, err := discordgo.New("Bot " + configuration.DiscordToken)
	if err != nil {
		l.FatalF("Error creating Discord session: %v", err)
	}
	defer dg.Close()
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		l.FatalF("Error opening connection with Discord: %v", err)
		return
	}

	// daemon.SdNotify(false, "READY=1")

	// Wait here until CTRL-C or other term signal is received.
	initBotCommands()
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	l.InfoF("Bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	l := LogInit("messageCreate-main.go")
	defer l.End()
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		// l.InfoF("Message is from bot itself, ignoring") // This is too talkative
		return
	}
	// bDM,:= ComesFromDM(s, m)
	// if err != nil {
	// 	log.Printf("Error checking for DM status")
	// }
	// if bDM {
	// 	log.Printf("Message was a DM!")
	// }
	// log.Printf("Session: %+v\nMessageCreate: %+v", s, m)
	// dumpPermissions(s, m)
	// userid := m.Author.ID
	// log.Printf("User %s is priv: %t", userid, isPriviledged(s, userid))

	// Split message between command and input
	// TODO: Make this smarter and less responses sent
	msg := strings.Split(m.Content, " ")
	resp := runCommand(s, m, msg)
	if resp != "" {
		if len(resp) > configuration.MaxMessageLength { // Max Discord message is 2000 characters, we need to break it up
			l.InfoF("Message too long, breaking it into multiple messages")
			chunked := chunkMessages(resp)
			for _, response := range chunked {
				s.ChannelMessageSend(m.ChannelID, response)
			}

			// splt := strings.Split(resp, "\n")
			// for _, line := range splt {
			// 	s.ChannelMessageSend(m.ChannelID, line)
			// }
		} else {
			s.ChannelMessageSend(m.ChannelID, resp)
		}
	}
}

func chunkMessages(r string) []string {
	full := strings.Split(r, " ")
	var response string
	var responses []string
	for _, word := range full {
		if (len(response) + len(word)) < configuration.MaxMessageLength {
			response = response + " " + word
		} else {
			responses = append(responses, response)
			response = ""
		}
	}
	return responses
}

func getUser(s *discordgo.Session) *discordgo.User {
	return s.State.User
}

// ComesFromDM returns true if a message comes from a DM channel
func ComesFromDM(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	l := LogInit("messageCreate-main.go")
	defer l.End()
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			l.WarnF("Failed to determind if this was a DM: %s", err.Error())
			return false
		}
	}

	return channel.Type == discordgo.ChannelTypeDM
}

func isPriviledged(s *discordgo.Session, userID string) bool {
	l := LogInit("messageCreate-main.go")
	defer l.End()
	guildID := configuration.GuildID
	l.InfoF("GuildID: %+v\nUserID: %+v", guildID, userID)
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		// if member, err = s.GuildMember(guildID, userID); err != nil {
		// 	return false, err
		// }
		l.ErrorF("Error: %s", err.Error())
	}
	l.InfoF("Member: %+v", member)
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			l.ErrorF("Error: %s", err.Error())
			return false
		}
		for _, cRole := range configuration.PrivRoles {
			l.InfoF("Crole: %v vs role.Name: %v", cRole, role.Name)
			if cRole == role.Name {
				l.InfoF("Role found, authorizing: %s == %s", cRole, role.Name)
				return true
			}
		}
	}
	return false
}

func connectDB() *sql.DB {
	db, err := sql.Open("mysql", "root:IGNOREME@/peq")
	if err != nil {
		panic(err)
	}
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func getResistsByMobName(name string, db *sql.DB) (NPC, error) {
	var npc NPC
	name = strings.ReplaceAll(name, " ", "_")
	// Prepare statement for reading data NAME LIKE "%Invisibility%Undead"
	name = fmt.Sprintf("%%%s%%", name)
	stmtOut, err := db.Prepare("SELECT id,NAME,Level,MR,CR,DR,FR,PR,Corrup,PhR,slow_mitigation,special_abilities,npcspecialattks FROM npc_types WHERE name LIKE ?")
	if err != nil {
		return NPC{}, err
	}
	defer stmtOut.Close()
	err = stmtOut.QueryRow(name).Scan(&npc.id, &npc.oName, &npc.level, &npc.mr, &npc.cr, &npc.dr, &npc.fr, &npc.pr, &npc.corrup, &npc.phr, &npc.slowMitigation, &npc.specialAbilities, &npc.npcSpecialAttacks)
	if err != nil {
		return NPC{}, err
	}
	return npc, nil
}

// NPC is a struct for holding mob data like resists, mitigation, special flags
type NPC struct {
	id                int
	name              string // is a clean name without underscores or #'s
	oName             string // contains the unclean name
	level             int
	mr                int
	cr                int
	dr                int
	fr                int
	pr                int
	corrup            int
	phr               int
	slowMitigation    int
	specialAbilities  string
	npcSpecialAttacks string
}

func getEvents(cal *calendar.Service, calID string, bDeleted bool, count int64, tFormat string) []Event {
	t := time.Now().Format(time.RFC3339)
	events, err := cal.Events.List(calID).ShowDeleted(false).SingleEvents(true).TimeMin(t).MaxResults(count).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	var es []Event
	for _, item := range events.Items {
		var e Event
		startDate := item.Start.DateTime
		if startDate == "" {
			startDate = item.Start.Date
		}
		tStart, err := time.Parse(time.RFC3339, startDate)
		endDate := item.End.DateTime
		if endDate == "" {
			endDate = item.End.Date
		}
		e.Start = tStart
		tEnd, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			fmt.Printf(err.Error())
		}
		e.End = tEnd
		e.Title = item.Summary
		e.Desc = item.Description
		es = append(es, e)
	}
	return es
}

// Event contains gcal info
type Event struct {
	Title string
	Start time.Time
	End   time.Time
	Desc  string
}
