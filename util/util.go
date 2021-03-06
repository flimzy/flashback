package util

import (
	"net/url"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

// JqmTargetUri determines the target URI based on a jQuery Mobile event 'ui' object
func JqmTargetUri(ui *js.Object) string {
	rawURL := ui.Get("toPage").String()
	if rawURL == "[object Object]" {
		rawURL = ui.Get("toPage").Call("jqmData", "url").String()
	}
	pageURL, _ := url.Parse(rawURL)
	pageURL.Path = strings.TrimPrefix(pageURL.Path, "/android_asset/www")
	pageURL.Host = ""
	pageURL.User = nil
	pageURL.Scheme = ""
	return pageURL.String()
}

/*
func UserDb() (*pouchdb.PouchDB, error) {
	userName := CurrentUser()
	if userName == "" {
		return nil, errors.New("Not logged in")
	}
	return pouchdb.New("user-" + userName), nil
}
*/

type reviewDoc struct {
	Id        string `json:"_id"`
	Rev       string `json:"_rev"`
	CurrentDb string `json:"CurrentDb"`
}

/*
func LogReview(r *fb.Review) error {
	db, err := ReviewsDb()
	if err != nil {
		return err
	}
	_, err = db.Put(r)
	if err != nil {
		return err
	}
	return nil
}

type DbList struct {
	Id  string   `json:"_id"`
	Rev string   `json:"_rev"`
	Dbs []string `json:"$dbs"`
}

func getReviewsDbList() (DbList, error) {
	var list DbList
	db, err := UserDb()
	if err != nil {
		return list, err
	}
	if err := db.Get("_local/ReviewsDbs", &list, pouchdb.Options{}); err != nil && pouchdb.IsNotExist(err) {
		return DbList{Id: "_local/ReviewsDbs"}, nil
	}
	return list, err
}

func setReviewsDbList(list DbList) error {
	if list.Id != "_local/ReviewsDbs" {
		return errors.Errorf("Invalid id '%s' for ReviewsDbs", list.Id)
	}
	log.Debugf("Setting list to: %v\n", list.Dbs)
	db, err := UserDb()
	if err != nil {
		return err
	}
	_, err = db.Put(list)
	return err
}

func ReviewsSyncDbs() (*repo.DB, error) {
	userName := CurrentUser()
	if userName == "" {
		return nil, errors.New("Not logged in")
	}
	list, err := getReviewsDbList()
	if err != nil {
		return nil, err
	}
	if len(list.Dbs) == 0 {
		// No reviews database, nothing to sync
		return nil, nil
	}
	dbName := list.Dbs[0]
	db := repo.NewDB(dbName)
	if len(list.Dbs) > 1 {
		log.Debugf("WARNING: More than one active reviews database!\n")
		return db, nil
	}
	var newDbPrefix string = "reviews-1-"
	if strings.HasPrefix(dbName, "reviews-1-") {
		newDbPrefix = "reviews-0-"
	}
	list.Dbs = append(list.Dbs, newDbPrefix+userName)
	if err := setReviewsDbList(list); err != nil {
		return nil, err
	}
	return db, nil
}

func ZapReviewsDb(db *repo.DB) error {
	info, err := db.Info()
	if err != nil {
		return err
	}
	list, err := getReviewsDbList()
	if err != nil {
		return err
	}
	if list.Dbs[0] != info.DBName {
		return errors.Errorf("Attempt to remove ReviewsDb '%s' not at head of list", info.DBName)
	}
	list.Dbs = list.Dbs[1:]
	if err := setReviewsDbList(list); err != nil {
		return err
	}
	return db.Destroy(pouchdb.Options{})
}

func ReviewsDb() (*pouchdb.PouchDB, error) {
	list, err := getReviewsDbList()
	if err != nil {
		return nil, err
	}
	if len(list.Dbs) == 0 {
		list.Dbs = []string{"reviews-0-" + CurrentUser()}
		err := setReviewsDbList(list)
		if err != nil {
			return nil, err
		}
	}
	dbName := list.Dbs[len(list.Dbs)-1]
	return pouchdb.New(dbName), nil
}
*/
