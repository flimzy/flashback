package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
)

type mockFile struct {
	body []byte
	err  error
}

var _ inputFile = &mockFile{}

func (f *mockFile) Bytes() ([]byte, error) { return f.body, f.err }

func TestImportFile(t *testing.T) {
	type ifTest struct {
		name string
		repo *Repo
		file inputFile
		err  string
	}
	tests := []ifTest{
		{
			name: "Not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "file read error",
			repo: &Repo{user: "bob"},
			file: &mockFile{err: errors.New("read error")},
			err:  "read error",
		},
		{
			name: "Invalid gzip data",
			repo: &Repo{user: "bob"},
			file: &mockFile{body: []byte("bogus data")},
			err:  "gzip: invalid header",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var msg string
			if err := test.repo.ImportFile(test.file); err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			if msg != "" {
				return
			}
		})
	}
}

func TestImport(t *testing.T) {
	type iTest struct {
		name     string
		repo     *Repo
		file     io.Reader
		expected *fb.Package
		err      string
	}
	id, _ := fb.NewDbID("bundle", []byte{1, 2, 3, 4})
	owner, _ := fb.NewDbID("user", []byte{1, 2, 3, 4, 5})
	tests := []iTest{
		{
			name: "Invalid JSON",
			repo: &Repo{user: "bob"},
			file: strings.NewReader("bogus data"),
			err:  "Unable to decode JSON: invalid character 'b' looking for beginning of value",
		},
		{
			name: "Not logged in",
			repo: &Repo{},
			file: strings.NewReader("{}"),
			err:  "not logged in",
		},
		{
			name: "Missing bundle db",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bob"); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			file: func() io.Reader {
				pkg := fb.Package{
					Bundle: &fb.Bundle{
						ID: id,
						Owner: &fb.User{
							ID: owner,
						},
					},
				}
				buf := &bytes.Buffer{}
				if err := json.NewEncoder(buf).Encode(pkg); err != nil {
					t.Fatal(err)
				}
				return buf
			}(),
			err: "bundleDB: database does not exist",
		},
		{
			name: "Invalid package",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bob"); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			file: func() io.Reader {
				pkg := fb.Package{
					Bundle: &fb.Bundle{
						ID: id,
						Owner: &fb.User{
							ID: owner,
						},
					},
					Cards: []*fb.Card{
						func() *fb.Card {
							c, err := fb.NewCard("theme-VGVzdCBUaGVtZQ", 0, "bundle-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0")
							if err != nil {
								t.Fatal(err)
							}
							return c
						}(),
					},
				}
				buf := &bytes.Buffer{}
				if err := json.NewEncoder(buf).Encode(pkg); err != nil {
					t.Fatal(err)
				}
				return buf
			}(),
			err: "card 'bundle-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0' found in package, but not in a deck",
		},
		{
			name: "Missing bundle",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bob"); err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), id.String()); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			file: func() io.Reader {
				pkg := fb.Package{}
				buf := &bytes.Buffer{}
				if err := json.NewEncoder(buf).Encode(pkg); err != nil {
					t.Fatal(err)
				}
				return buf
			}(),
			err: "invalid bundle",
		},
		{
			name: "Valid",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bob"); err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), id.String()); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			file: func() io.Reader {
				pkg := fb.Package{
					Bundle: &fb.Bundle{
						ID: id,
						Owner: &fb.User{
							ID: owner,
						},
					},
				}
				buf := &bytes.Buffer{}
				if err := json.NewEncoder(buf).Encode(pkg); err != nil {
					t.Fatal(err)
				}
				return buf
			}(),
			expected: &fb.Package{
				Bundle: &fb.Bundle{
					ID:  id,
					Rev: func() *string { x := "1"; return &x }(),
					Owner: &fb.User{
						ID: owner,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var msg string
			err := test.repo.Import(context.Background(), test.file)
			if err != nil {
				msg = err.Error()
			}
			if test.err != msg {
				t.Errorf("Unexpected error: %s", msg)
			}
			if err != nil {
				return
			}
			udb, err := test.repo.userDB(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			checkBundle(t, udb, test.expected.Bundle)
			bdb, err := test.repo.bundleDB(context.Background(), test.expected.Bundle)
			if err != nil {
				t.Fatal(err)
			}
			checkBundle(t, bdb, test.expected.Bundle)
			// row, err := udb.Get(context.Background(), test.expected.Bundle.ID.String())
			// if err != nil {
			// 	t.Fatalf("failed to refectch bundle from userdb: %s", err)
			// }
			// bundle := &fb.Bundle{}
			// if e := row.ScanDoc(&bundle); e != nil {
			// 	t.Fatal(e)
			// }
			// revParts := strings.Split(*bundle.Rev, "-")
			// bundle.Rev = &revParts[0]
			// if d := diff.AsJSON(test.expected.Bundle, bundle); d != "" {
			// 	t.Error(d)
			// }
		})
	}
}

/*
var UUID = []byte{0xD1, 0xC9, 0x58, 0x7D, 0x88, 0xDF, 0x4A, 0x65, 0x89, 0x23, 0xF7, 0x3C, 0xDF, 0x6D, 0x1D, 0x70}

func BenchmarkSaveCard(b *testing.B) {
	u, err := fb.NewUser(uuid.UUID(UUID), "testuser")
	if err != nil {
		panic(err)
	}
	user := &User{u}
	client, _ := kivik.New(context.TODO(), "pouch", "")
	if err = client.DestroyDB(context.TODO(), u.ID.String()); err != nil {
		panic(err)
	}
	db, err := user.DB()
	if err != nil {
		panic(err)
	}
	err = db.CreateIndex(context.TODO(), "", "", map[string]interface{}{
		"fields": []string{"due", "created", "type"},
	})
	if err != nil {
		panic(err)
	}
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
*/
