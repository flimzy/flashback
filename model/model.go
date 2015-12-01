package model

import (
	"golang.org/x/net/context"

	"github.com/flimzy/go-pouchdb"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback/clientstate"
)

func UserDB(ctx context.Context) (string,error) {
	state := ctx.Value("AppState").(*clientstate.State)
	return "user-" + state.CurrentUser,nil
}

type Deck struct {
	ID          string
	Description string
}

func GetDecksList(ctx context.Context) ([]*Deck,error) {
	dbName,err := UserDB(ctx)
	if err != nil {
		return nil,err
	}
	db := pouchdb.New(dbName)
	var doc struct {
		Total	int		`json:"total_rows"`
		Rows []struct{
			Deck Deck	`json:"value"`
		} `json:"rows"`
	}
	err = db.Query("index/decks", &doc, pouchdb.Options{})
	if err != nil {
		return nil,err
	}
	console.Log("%v", doc)
	decks := make([]*Deck,len(doc.Rows))
	for i,row := range doc.Rows {
		decks[i] = &row.Deck
	}
	return decks,nil
}
