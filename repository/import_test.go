package repo

import (
	"fmt"
	"testing"

	pouchdb "github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/pborman/uuid"

	"github.com/FlashbackSRS/flashback-model"
)

var UUID = []byte{0xD1, 0xC9, 0x58, 0x7D, 0x88, 0xDF, 0x4A, 0x65, 0x89, 0x23, 0xF7, 0x3C, 0xDF, 0x6D, 0x1D, 0x70}

func BenchmarkSaveCard(b *testing.B) {
	u, err := fb.NewUser(uuid.UUID(UUID), "testuser")
	if err != nil {
		panic(err)
	}
	user := &User{u}
	db, err := user.DB()
	if err != nil {
		panic(err)
	}
	db.Destroy(pouchdb.Options{})
	db, err = user.DB()
	if err != nil {
		panic(err)
	}
	_ = db.CreateIndex(find.Index{
		Fields: []string{"due", "created", "type"},
	})
	cards := make([]*fb.Card, b.N)
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("card-bundle.%x.0", i)
		card, _ := fb.NewCard("themefoo", 0, id)
		cards[i] = card
	}
	b.ResetTimer()
	for _, card := range cards {
		if err := db.Save(card); err != nil {
			panic(err)
		}
	}
}