package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const configPath = "config.json"
const failPath = "C:\\"

var configuration Configuration

// Configuration stores all our user defined variables
type Configuration struct {
	LogLevel                int       `json:"LogLevel"`                    // 0=Off,1=Error,2=Warn,3=Info,4=Debug
	LogPath                 string    `json:"LogPath"`                     // Where to write logs to
	DiscordToken            string    `json:"DiscordToken"`                // Discord Bot Token for Authentication
	AccessToken             string    `json:"access_token"`                // Google Access Token
	TokenType               string    `json:"token_type"`                  // Google Token Type
	RefreshToken            string    `json:"refresh_token"`               // Google Refresh Token
	Expiry                  time.Time `json:"expiry"`                      // Google Expiration Date
	ClientID                string    `json:"client_id"`                   // Google Client ID
	ProjectID               string    `json:"project_id"`                  // Google Project ID
	AuthURI                 string    `json:"auth_uri"`                    // Google Auth URI
	TokenURI                string    `json:"token_uri"`                   // Google Token URI
	AuthProviderx509CertURL string    `json:"auth_provider_x509_cert_url"` // Google Cert URL
	ClientSecret            string    `json:"client_secret"`               // Google Client Secret
	RedirectURIs            []string  `json:"redirect_uris"`               // Google Redirect URIs
	// --------
	DKPSheetURL                string   `json:"DKPSheetURL"`                // String after https://docs.google.com/spreadsheets/d/ and before /edit
	DKPSheetName               string   `json:"DKPSheetName"`               // Sheet Name for DKP
	DKPSheetClassCol           int      `json:"DKPSheetClassRow"`           // Column for player class BASE 0
	DKPSheetRankCol            int      `json:"DKPSheetRankCol"`            // Column for player rank BASE 0
	DKPSheetNameCol            int      `json:"DKPSheetNameCol"`            // Column for player name BASE 0
	DKPSheetLevelCol           int      `json:"DKPSheetLevelCol"`           // Column for player level BASE 0
	DKPSheetLastRaidCol        int      `json:"DKPSheetLastRaidCol"`        // Column for player last raid attended BASE 0
	DKPSheetAttendanceCol      int      `json:"DKPSheetAttendanceCol"`      // Column for player attendance ratio Base 0
	DKPSheetDKPCol             int      `json:"DKPSheetDKPCol"`             // Column for player DKP
	DKPSummarySheetName        string   `json:"DKPSummarySheetName"`        // Sheet Name for DKP Summary
	DKPSummarySheetDateCol     int      `json:"DKPSummarySheetDateCol"`     // Column for DKP Summary Date
	DKPSummarySheetPlayerCol   int      `json:"DKPSummarySheetPlayerCol"`   // Column for DKP Summary player
	DKPSummarySheetDKPDescCol  int      `json:"DKPSummarySheetDKPDescCol"`  // Column for DKP Description
	DKPSummarySheetDKPCol      int      `json:"DKPSummarySheetDKPCol"`      // Column for DKP Summary DKP
	DKPSRosterSheetName        string   `json:"DKPSRosterSheetName"`        // Sheet Name for DKP Rsoter
	DKPSRosterSheetPlayerCol   int      `json:"DKPSRosterSheetPlayerCol"`   // Column for DKP Summary Date
	DKPSRosterSheetLevelCol    int      `json:"DKPSRosterSheetLevelCol"`    // Column for DKP Summary player
	DKPSRosterSheetClassCol    int      `json:"DKPSRosterSheetClassCol"`    // Column for DKP Description
	DKPSRosterSheetRankCol     int      `json:"DKPSRosterSheetRankCol"`     // Column for DKP Summary DKP
	DKPSRosterSheetJoinDateCol int      `json:"DKPSRosterSheetJoinDateCol"` // Column for DKP Summary DKP
	SpellSheet                 string   `json:"SpellSheet"`                 // Sheet Name for Spells
	SpellSheetHeaderRow        int      `json:"SpellSheetHeaderRow"`        // Row # for Spell Sheet's header containing player named BASE 0
	SpellSheetSpellCol         int      `json:"SpellSheetSpellCol"`         // Column for spell names
	RulesSheetName             string   `json:"RulesSheetName"`             // Sheet Name for Rules
	GuildID                    string   `json:"GuildID"`                    // Discord Guild ID
	PrivRoles                  []string `json:"PrivRoles"`                  // Array of roles that can run piviledged commands Exact String Match
	NoPrivResponse             string   `json:"NoPrivResponse"`             // Response given if the user attempts a priv command unpriv
	MaxMessageLength           int      `json:"MaxMessageLength"`           // Max Discord message length (2000)
	KronoAPIURL                string   `json:"KronoAPIURL"`                // Aradune Auctions krono API URL
	CommDKPCommand             string   `json:"CommDKPCommand"`             // String to trigger for DKP Command
	CommDKPHelp                string   `json:"CommDKPHelp"`                // Aradune Auctions krono API URL
	CommDKPDMOnly              bool     `json:"CommDKPDMOnly"`              // Aradune Auctions krono API URL
	CommDKPPriv                bool     `json:"CommDKPPriv"`                // Aradune Auctions krono API URL
	CommDKPHidden              bool     `json:"CommDKPHidden"`              // Aradune Auctions krono API URL
	CommRaidSummaryCommand     string   `json:"CommRaidSummaryCommand"`     // Aradune Auctions krono API URL
	CommRaidSummaryHelp        string   `json:"CommRaidSummaryHelp"`        // Aradune Auctions krono API URL
	CommRaidSummaryDMOnly      bool     `json:"CommRaidSummaryDMOnly"`      // Aradune Auctions krono API URL
	CommRaidSummaryPriv        bool     `json:"CommRaidSummaryPriv"`        // Aradune Auctions krono API URL
	CommRaidSummaryHidden      bool     `json:"CommRaidSummaryHidden"`
	CommHelpCommand            string   `json:"CommHelpCommand"` // Aradune Auctions krono API URL
	CommHelpHelp               string   `json:"CommHelpHelp"`    // Aradune Auctions krono API URL
	CommHelpDMOnly             bool     `json:"CommHelpDMOnly"`  // Aradune Auctions krono API URL
	CommHelpPriv               bool     `json:"CommHelpPriv"`    // Aradune Auctions krono API URL
	CommHelpHidden             bool     `json:"CommHelpHidden"`
	CommDBRCommand             string   `json:"CommDBRCommand"` // Aradune Auctions krono API URL
	CommDBRHelp                string   `json:"CommDBRHelp"`    // Aradune Auctions krono API URL
	CommDBRDMOnly              bool     `json:"CommDBRDMOnly"`  // Aradune Auctions krono API URL
	CommDBRPriv                bool     `json:"CommDBRPriv"`    // Aradune Auctions krono API URL
	CommDBRHidden              bool     `json:"CommDBRHidden"`
	CommKronoCommand           string   `json:"CommKronoCommand"` // Aradune Auctions krono API URL
	CommKronoHelp              string   `json:"CommKronoHelp"`    // Aradune Auctions krono API URL
	CommKronoDMOnly            bool     `json:"CommKronoDMOnly"`  // Aradune Auctions krono API URL
	CommKronoPriv              bool     `json:"CommKronoPriv"`    // Aradune Auctions krono API URL
	CommKronoHidden            bool     `json:"CommKronoHidden"`
	CommSpellCommand           string   `json:"CommSpellCommand"` // Aradune Auctions krono API URL
	CommSpellHelp              string   `json:"CommSpellHelp"`    // Aradune Auctions krono API URL
	CommSpellDMOnly            bool     `json:"CommSpellDMOnly"`  // Aradune Auctions krono API URL
	CommSpellPriv              bool     `json:"CommSpellPriv"`    // Aradune Auctions krono API URL
	CommSpellHidden            bool     `json:"CommSpellHidden"`
	CommGiveSpellCommand       string   `json:"CommGiveSpellCommand"` // Aradune Auctions krono API URL
	CommGiveSpellHelp          string   `json:"CommGiveSpellHelp"`    // Aradune Auctions krono API URL
	CommGiveSpellDMOnly        bool     `json:"CommGiveSpellDMOnly"`  // Aradune Auctions krono API URL
	CommGiveSpellPriv          bool     `json:"CommGiveSpellPriv"`    // Aradune Auctions krono API URL
	CommGiveSpellHidden        bool     `json:"CommGiveSpellHidden"`
	CommRulesCommand           string   `json:"CommRulesCommand"` // Aradune Auctions krono API URL
	CommRulesHelp              string   `json:"CommRulesHelp"`    // Aradune Auctions krono API URL
	CommRulesDMOnly            bool     `json:"CommRulesDMOnly"`  // Aradune Auctions krono API URL
	CommRulesPriv              bool     `json:"CommRulesPriv"`    // Aradune Auctions krono API URL
	CommRulesHidden            bool     `json:"CommRulesHidden"`
	CommDKPClassCommand        string   `json:"CommDKPClassCommand"` // Aradune Auctions krono API URL
	CommDKPClassHelp           string   `json:"CommDKPClassHelp"`    // Aradune Auctions krono API URL
	CommDKPClassDMOnly         bool     `json:"CommDKPClassDMOnly"`  // Aradune Auctions krono API URL
	CommDKPClassPriv           bool     `json:"CommDKPClassPriv"`    // Aradune Auctions krono API URL
	CommDKPClassHidden         bool     `json:"CommDKPClassHidden"`
}

func init() {
	readConfig()
	log.Printf("Configuration loaded:\n %+v\n", configuration)
}

func readConfig() error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println(err)
	}
	if _, err := os.Stat(dir + "/" + configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		dir, _ = os.Getwd()
	}
	if _, err := os.Stat(dir + "/" + configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		dir = failPath
	}
	if _, err := os.Stat(dir + "/" + configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		log.Fatal(err)
	}
	file, err := os.OpenFile(dir+"/"+configPath, os.O_RDONLY, 0444)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return err
	}
	return nil
}

func saveConfig() error {
	marshalledConfig, _ := json.MarshalIndent(configuration, "", "\t")
	err := ioutil.WriteFile(configPath, marshalledConfig, 0644)
	if err != nil {
		return err
	}
	log.Printf("Config Saved to %s\n", configPath)
	return nil
}
