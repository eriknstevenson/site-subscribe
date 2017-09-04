package sitesubscribe

import (
	"crypto/md5"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	// sqlite3 driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// gorm Models

type User struct {
	gorm.Model
	Name  string `gorm:"not null"`
	Email string `gorm:"not null;unique"`
	Sites []Site
}

type Site struct {
	gorm.Model
	Name            string `gorm:"not null;"`
	URL             string `gorm:"not null;unique"`
	ContentSelector string
	ContentHashes   []ContentHash
	UserID          uint
}

type ContentHash struct {
	gorm.Model
	SiteID uint
	Hash   string `gorm:"not null"`
}

// New sets up and returns a siteSubscribe object.
func connectToDB() *gorm.DB {

	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&User{}, &Site{}, &ContentHash{})

	return db
}

// Close cleans up and should be called when the SiteSubscribe object is no longer required.
func closeDB(db *gorm.DB) {
	db.Close()
}

// AddUser creates a new user.
func AddUser(userName, userEmail string) {
	db := connectToDB()
	defer db.Close()

	db.Create(&User{Name: userName, Email: userEmail})
}

// AddSite subscribes a user to a site.
func AddSite(userEmail string, site Site) {
	db := connectToDB()
	defer db.Close()

	var user User
	db.Where(&User{Email: userEmail}).First(&user)

	user.Sites = append(user.Sites, site)
	db.Save(user)
}

// Update checks all sites for changes, and sends out notifications as necessary.
func Update(sendgridKey string) {

	db := connectToDB()
	defer db.Close()

	sendgridClient := sendgrid.NewSendClient(sendgridKey)

	var users []User
	db.Find(&users)

	for _, user := range users {

		var userSites []Site
		db.Model(user).Related(&userSites)

		var updatedSites []Site

		for _, site := range userSites {

			doc, err := goquery.NewDocument(site.URL)
			if err != nil {
				log.Printf("could not get page %s: %v", site.URL, err)
				continue
			}

			content, err := doc.Find(func(s string) string {
				if s == "" {
					return "body"
				}
				return s
			}(site.ContentSelector)).Html()
			if err != nil {
				log.Printf("could get retrieve content from %s: %v", site.URL, err)
			}

			log.Printf("Downloaded content from: %s", site.URL)

			hashArray := md5.Sum([]byte(content))
			newHash := ContentHash{
				Hash: fmt.Sprintf("%x", hashArray),
			}

			log.Printf("Hashed content: %v", newHash)

			var lastHash ContentHash
			db.Where(&ContentHash{SiteID: site.ID}).Last(&lastHash)

			if lastHash.Hash != "" && lastHash.Hash != newHash.Hash {
				updatedSites = append(updatedSites, site)
			}

			site.ContentHashes = append(site.ContentHashes, newHash)
			db.Save(site)

		}

		if len(updatedSites) > 0 {
			log.Printf("Changes detected on %d sites, notifying user %s", len(updatedSites), user.Name)
			notify(sendgridClient, user, updatedSites)
		}
	}
}

func notify(sendgridClient *sendgrid.Client, user User, sites []Site) {
	from := mail.NewEmail("Site Subscribe", "eriknstevenson@gmail.com")
	if len(sites) == 0 {
		return
	}
	var subject string

	var siteNames []string
	for _, site := range sites {
		siteNames = append(siteNames, site.Name)
	}
	combinedSiteNames := strings.Join(siteNames, ", ")
	subject = fmt.Sprintf("%s Updated", combinedSiteNames)

	var plainTextContent string
	plainTextContent = fmt.Sprintf("Dear %s,\nAn update was recently made to %s.\n", user.Name, func(n int) string {
		if n == 1 {
			return fmt.Sprintf("the %s website", sites[0].Name)
		}
		return "a few of the sites you're subscribed to"
	}(len(sites)))

	var htmlContent string
	htmlContent = fmt.Sprintf("<h2>%s Updated</h2>Dear %s,<br>An update was recently made to %s.\n", combinedSiteNames, user.Name, func(n int) string {
		if n == 1 {
			return fmt.Sprintf("the <a href=%s>%s</a> website", sites[0].URL, sites[0].Name)
		}
		return "a few of the sites you're subscribed to"
	}(len(sites)))

	htmlContent += "<ul>"

	for _, site := range sites {
		plainTextContent += fmt.Sprintf("    %s - %s\n", site.Name, site.URL)
		htmlContent += fmt.Sprintf("<li><a href=%s>%s</a></li>", site.URL, site.Name)
	}

	htmlContent += "</ul>"

	plainTextContent += fmt.Sprintf("-------\nTo unsubscribe from these notifications, click here.\n")
	htmlContent += fmt.Sprintf("<hr><small>To unsubscribe from these notifications, click <a href=\"\">here</a>.</small>")

	to := mail.NewEmail(user.Name, user.Email)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	response, err := sendgridClient.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
