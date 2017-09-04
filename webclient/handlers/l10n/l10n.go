package l10n_handler

import (
	"net/url"
	"sync"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/l10n"
)

var jQuery = jquery.NewJQuery

const langTagAttr = "data-lt"
const localeAttr = "data-locale"

// Init initializes the localization engine.
func Init() *l10n.Set {
	fetch := fetchTranslations
	if cordova.IsMobile() {
		fetch = fetchTranslationsCordova
	}
	set, err := l10n.New(preferredLanguages, fetch)
	if err != nil {
		panic(err)
	}
	return set
}

func preferredLanguages() ([]string, error) {
	var langs []string
	nav := js.Global.Get("navigator")
	if cordova.IsMobile() {
		var wg sync.WaitGroup
		wg.Add(1)
		nav.Get("globalization").Call("getPreferredLanguage",
			func(l *js.Object) {
				langs = append(langs, l.Get("value").String())
				wg.Done()
			}, func(e *js.Object) {
				// ignore any error
				wg.Done()
			})
		wg.Wait()
	}
	if languages := nav.Get("languages"); languages != js.Undefined {
		for i := 0; i < languages.Length(); i++ {
			langs = append(langs, languages.Index(i).String())
		}
	}
	langs = append(langs, nav.Get("language").String())
	return langs, nil
}

// LocalizePage localizes all language tags on the page.
func LocalizePage(set *l10n.Set, h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		go localize(set)
		return h.HandleEvent(event, ui, p)
	})
}

func localize(set *l10n.Set) {
	T, err := set.Tfunc()
	if err != nil {
		panic(err)
	}
	elements := jQuery("[" + langTagAttr + "]").Not("[" + localeAttr + "='" + set.Locale + "']")
	if elements.Length == 0 {
		log.Debugf("Nothing to localize, early exiting\n")
		return
	}
	log.Debugf("I think I have %d items to localize\n", elements.Length)
	for i := 0; i < elements.Length; i++ {
		elem := elements.Get(i)
		text := elem.Call("getAttribute", langTagAttr).String()
		if translated := T(text); translated != text {
			elem.Set("innerHTML", translated)
		} else {
			log.Debugf("%d. %s = %s\n", i, text, translated)
		}
	}
}
