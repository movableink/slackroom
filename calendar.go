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
	"strings"
	"time"
)

type Calendar struct {
	service *calendar.Service
	roomMap map[string]string
}

func NewCalendarService() *Calendar {
	ctx := context.Background()

	b, err := ioutil.ReadFile("/home/deploy/apps/slackroom/client_secret.json")

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)

	cacheFile, err := tokenCacheFile()

	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}

	client := getClient(ctx, config)

	service, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
		return nil
	}

	roomMap := make(map[string]string)
	listRes, err := service.CalendarList.List().Fields("items").Do()

	if err != nil {
		log.Fatalf("Unable to find calendars %v", err)
	}

	log.Printf("%s", listRes.Items)

	for _, item := range listRes.Items {
		if strings.Contains(item.Summary, "MI Room") {
			roomMap[item.Id] = item.Summary
		}
	}

	return &Calendar{service: service, roomMap: roomMap}
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
	log.Printf("retrieving new token")

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

// func (srv *Calendar) QuickBook(room string) string {
// 	res := srv.service.Events.QuickAdd(room, "booked by eric")
// 	log.Printf("Unable to book", res)
// 	return "booked"
// }

func (srv *Calendar) GetAvaliableRooms() string {
	var requestItem []*calendar.FreeBusyRequestItem
	log.Printf("%s", srv.roomMap)

	for k, _ := range srv.roomMap {
		item := &calendar.FreeBusyRequestItem{Id: k}
		requestItem = append(requestItem, item)
	}

	if len(requestItem) > 0 {
		now := time.Now()
		later := now.Add(15 * time.Minute)

		request := &calendar.FreeBusyRequest{
			TimeMin:  now.Format(time.RFC3339),
			TimeMax:  later.Format(time.RFC3339),
			TimeZone: "EST",
			Items:    requestItem,
		}

		res, err := srv.service.Freebusy.Query(request).Do()

		if err != nil {
			log.Fatalf("Unable to query Freebusy %v", err)
		}

		var avaliable string
		for k, c := range res.Calendars {
			if len(c.Busy) == 0 {
				avaliable = avaliable + " " + (srv.roomMap)[k]
			}
		}

		if len(avaliable) > 0 {
			return string(avaliable)
		}
	}
	return "none"
}
