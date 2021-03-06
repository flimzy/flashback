package model

import (
	"context"
	"errors"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

func TestBuryInterval(t *testing.T) {
	tests := []struct {
		name     string
		bury     fb.Interval
		interval fb.Interval
		new      bool
		expected fb.Interval
	}{
		{
			name:     "old card",
			bury:     10 * fb.Day,
			interval: 20 * fb.Day,
			new:      false,
			expected: 4 * fb.Day,
		},
		{
			name:     "new card",
			bury:     10 * fb.Day,
			interval: 20 * fb.Day,
			new:      true,
			expected: 7 * fb.Day,
		},
		{
			name:     "1 day",
			bury:     10 * fb.Day,
			interval: 1 * fb.Day,
			new:      false,
			expected: 1 * fb.Day,
		},
		{
			name:     "maxiaml burial",
			bury:     20 * fb.Day,
			interval: 90 * fb.Day,
			new:      false,
			expected: 10 * fb.Day,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buryInterval(test.bury, test.interval, test.new)
			if result != test.expected {
				t.Errorf("%s / %s / %t:\n\tExpected: %s\n\t  Actual: %s\n", test.bury, test.interval, test.new, test.expected, result)
			}
		})
	}
}

func TestFetchRelatedCards(t *testing.T) {
	tests := []struct {
		name     string
		db       allDocer
		cardID   string
		expected []*fb.Card
		err      string
	}{
		{
			name:   "db error",
			db:     &mockAllDocer{err: errors.New("db error")},
			cardID: "card-foo.bar.0",
			err:    "db error",
		},
		{
			name:   "iteration error",
			db:     &mockAllDocer{rows: &mockRows{err: errors.New("db error")}},
			cardID: "card-foo.bar.0",
			err:    "db error",
		},
		{
			name: "invalid json",
			db: &mockAllDocer{
				rows: &mockRows{
					rows: []string{
						`{"_id":"card-foo.bar.1", "created":"2017-01-01T01:01:01Z", "modified":12345, "model": "theme-Zm9v/0"}`,
					},
				},
			},
			cardID: "card-foo.bar.0",
			err:    `scan doc: parsing time "12345" as ""2006-01-02T15:04:05Z07:00"": cannot parse "12345" as """`,
		},
		{
			name: "success",
			db: &mockAllDocer{
				rows: &mockRows{
					rows: []string{
						`{"_id":"card-foo.bar.0", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "model": "theme-Zm9v/0"}`,
						`{"_id":"card-foo.bar.1", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "model": "theme-Zm9v/0"}`,
					},
				},
			},
			cardID: "card-foo.bar.0",
			expected: []*fb.Card{
				{
					ID:       "card-foo.bar.1",
					ModelID:  "theme-Zm9v/0",
					Created:  parseTime(t, "2017-01-01T01:01:01Z"),
					Modified: parseTime(t, "2017-01-01T01:01:01Z"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := fetchRelatedCards(context.Background(), test.db, test.cardID)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

type buryClient struct {
	kivikClient
	db kivikDB
}

var _ kivikClient = &buryClient{}

func (c *buryClient) DB(_ context.Context, _ string, _ ...kivik.Options) (kivikDB, error) {
	return c.db, nil
}

func TestBuryRelatedCards(t *testing.T) {
	tests := []struct {
		name string
		repo *Repo
		card *fb.Card
		err  string
	}{
		{
			name: "not logged in",
			repo: &Repo{},
			card: &fb.Card{ID: "card-foo.bar.0"},
			err:  "not logged in",
		},
		{
			name: "fetch error",
			repo: &Repo{user: "bob",
				local: &buryClient{
					db: &mockAllDocer{
						err: errors.New("db error"),
					},
				},
			},
			card: &fb.Card{ID: "card-foo.bar.0"},
			err:  "db error",
		},
		{
			name: "no related cards",
			repo: &Repo{user: "bob",
				local: &buryClient{db: &mockAllDocer{
					rows: &mockRows{},
				}}},
			card: &fb.Card{ID: "card-foo.bar.0"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.repo.BuryRelatedCards(context.Background(), test.card)
			checkErr(t, test.err, err)
		})
	}
}

func TestSetBurials(t *testing.T) {
	tests := []struct {
		name     string
		interval fb.Interval
		cards    []*fb.Card
		expected []*fb.Card
	}{
		{
			name:     "no cards",
			cards:    []*fb.Card{},
			expected: []*fb.Card{},
		},
		{
			name:     "two cards",
			interval: fb.Interval(24 * time.Hour),
			cards: []*fb.Card{
				{}, // new
				{
					ReviewCount: 1,
					Interval:    fb.Interval(24 * time.Hour),
				}, // Minimal burial
				{
					ReviewCount: 1,
					BuriedUntil: fb.Due(parseTime(t, "2018-01-01T00:00:00Z")),
				}, // Should not be re-buried
			},
			expected: []*fb.Card{
				{BuriedUntil: fb.Due(parseTime(t, "2017-01-08T00:00:00Z"))},
				{
					ReviewCount: 1,
					Interval:    fb.Interval(24 * time.Hour),
					BuriedUntil: fb.Due(parseTime(t, "2017-01-02T00:00:00Z")),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := setBurials(test.interval, test.cards)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
