package model

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
)

func TestSaveBundle(t *testing.T) {
	type sbTest struct {
		name     string
		repo     *Repo
		bundle   *fb.Bundle
		expected map[string]interface{}
		err      string
	}
	id := fb.EncodeDBID("bundle", []byte{1, 2, 3, 4})
	tests := []sbTest{
		{
			name:   "not logged in",
			repo:   &Repo{},
			bundle: &fb.Bundle{ID: id},
			err:    "not logged in",
		},
		{
			name:   "invalid bundle",
			repo:   &Repo{user: "bob"},
			bundle: &fb.Bundle{},
			err:    "invalid bundle: id required",
		},
		{
			name: "user db does not exist",
			repo: &Repo{
				user:  "bob",
				local: &mockClient{err: errors.New("database does not exist")},
			},
			bundle: &fb.Bundle{ID: id, Created: now(), Modified: now(), Owner: "mjxwe"},
			err:    "userDB: database does not exist",
		},
		{
			name: "success",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "user-mjxwe"); err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), id); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					local: local,
					user:  "mjxwe",
				}
			}(),
			bundle: &fb.Bundle{ID: id, Owner: "mjxwe", Created: now(), Modified: now()},
			expected: map[string]interface{}{
				"_id":      id,
				"type":     "bundle",
				"_rev":     "1",
				"owner":    "mjxwe",
				"created":  now(),
				"modified": now(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.repo.SaveBundle(context.Background(), test.bundle)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
				return
			}
			if err != nil {
				return
			}
			udb, err := test.repo.userDB(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			bdb, err := test.repo.bundleDB(context.Background(), test.bundle)
			if err != nil {
				t.Fatal(err)
			}
			checkDoc(t, udb, test.expected)
			checkDoc(t, bdb, test.expected)
		})
	}
}

func checkDoc(t *testing.T, db getter, doc interface{}) {
	var docID string
	switch b := doc.(type) {
	case map[string]interface{}:
		docID = b["_id"].(string)
	case *fb.Bundle:
		docID = b.ID
	default:
		x, err := json.Marshal(doc)
		if err != nil {
			panic(err)
		}
		var result struct {
			ID string `json:"_id"`
		}
		if e := json.Unmarshal(x, &result); e != nil {
			panic(e)
		}
		docID = result.ID
	}
	row, err := db.Get(context.Background(), docID)
	if err != nil {
		t.Errorf("failed to fetch %s: %s", docID, err)
		return
	}
	var result map[string]interface{}
	if err := row.ScanDoc(&result); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(result["_rev"].(string), "-")
	result["_rev"] = parts[0]
	delete(result, "_attachments")
	if d := diff.AsJSON(doc, result); d != nil {
		t.Error(d)
	}
}
