package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type CalendarClient interface {
	QuickBook()
	SetAvaliableRooms()
}

type Service struct {
	calendar *calendar.Service
}

func NewCalendarService() *Service {
	ctx := context.Background()

	b, err := ioutil.ReadFile("/Users/ericcook/go/src/github.com/movableink/slackroom/bin/client_secret.json")

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)

	cacheFile, err := tokenCacheFile()
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}

	client := getClient(ctx, config)

	calendar, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
		return nil
	}

	return &Service{calendar: calendar}
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("slackroom.json")), err
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}
func (srv *Service) QuickBook(room string) string {
	res := srv.calendar.Events.QuickAdd(room, "booked by eric")
	log.Printf("Unable to book", res)
	return "booked"
}

func (srv *Service) GetAvaliableRooms() string {
	listRes, err := srv.calendar.CalendarList.List().Fields("items(id)").Do()
	if err != nil {
		log.Fatalf("Unable to find calendars %v", err)
	}

	var calIds []*calendar.FreeBusyRequestItem

	for _, v := range listRes.Items {
		item := &calendar.FreeBusyRequestItem{Id: v.Id}
		calIds = append(calIds, item)
	}

	if len(calIds) > 0 {
		now := time.Now()
		later := now.Add(30 * time.Minute)

		request := &calendar.FreeBusyRequest{
			TimeMin:  now.Format(time.RFC3339),
			TimeMax:  later.Format(time.RFC3339),
			TimeZone: "EST",
			Items:    calIds,
		}

		res, err := srv.calendar.Freebusy.Query(request).Do()

		if err != nil {
			log.Fatalf("Unable to query Freebusy %v", err)
		}

		var avaliable string
		for k, c := range res.Calendars {
			if len(c.Busy) == 0 {
				avaliable = avaliable + " " + k + "\n"
			}
		}

		if len(avaliable) > 0 {
			return string(avaliable)
		}
	}
	return "no"
}
