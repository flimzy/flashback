package sync

import (
	"golang.org/x/net/context"

	"github.com/flimzy/flashback/clientstate"
	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/go-pouchdb"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func init() {
	pages.Register("/sync.html", "pagecontainerbeforetransition", BeforeTransition)
}

func BeforeTransition(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("sync BEFORE")

	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#syncnow", container).On("click", func() {
			console.Log("Attempting to sync something...")
			go DoSync(ctx)
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return pages.Return()
}

func DoSync(ctx context.Context) {
	state := ctx.Value("AppState").(*clientstate.State)
	host := ctx.Value("couchhost").(string)
	dbName := "user-" + state.CurrentUser
	user_db := pouchdb.New(host + "/" + dbName)
	skeleton_db := pouchdb.New(host + "/user-skeleton")
	result, err := pouchdb.Replicate(skeleton_db, user_db, pouchdb.Options{})
	if err != nil {
		console.Log("Skel error =  %j", err)
		return
	}
	console.Log("skel result = %j", result)
	local_db := pouchdb.New(dbName)
	result, err = pouchdb.Replicate(user_db, local_db, pouchdb.Options{})
	if err != nil {
		console.Log("error = %j", err)
		return
	}
	console.Log("result = %j", result)
}
