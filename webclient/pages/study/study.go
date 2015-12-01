package study

import (
	"golang.org/x/net/context"

	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/flashback/model"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func init() {
	pages.Register("/study.html", "pagecontainerbeforetransition", BeforeTransition)
}

func BeforeTransition(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("study BEFORE")

	go func() {
		container := jQuery(":mobile-pagecontainer")
		decks,err := model.GetDecksList(ctx)
		if err != nil {
			console.Log("Error fetching decks: %s", err)
		}
		console.Log("decks = %j", decks)

		modelDeck := jQuery(".deck", container)
//		deckParent := modelDeck.Call("parent")
		for _,deck := range decks {
			newDeck := modelDeck.Call("clone")
			newDeck.SetHtml( deck.Description )
			newDeck.InsertAfter( modelDeck )
		}
		modelDeck.Remove()

		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()
	return pages.Return()
}
