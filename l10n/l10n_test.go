package l10n

import (
	"errors"
	"testing"
)

const testID = "foo"

type mockSource struct {
	langs func() ([]string, error)
	fetch func(string) ([]byte, error)
}

func (s *mockSource) Languages() ([]string, error) {
	return s.langs()
}

func (s *mockSource) FetchLanguage(lang string) ([]byte, error) {
	return s.fetch(lang)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		src      Source
		err      string
		initErr  string
		expected string
	}{
		{
			name: "No langs callback",
			err:  "src required",
		},
		{
			name: "No preference",
			src: &mockSource{
				langs: func() ([]string, error) { return nil, nil },
				fetch: func(string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Foo"}]`), nil },
			},
			expected: "Foo",
		},
		{
			name: "Preference is default",
			src: &mockSource{
				langs: func() ([]string, error) { return []string{"en_US"}, nil },
				fetch: func(_ string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Foo"}]`), nil },
			},
			expected: "Foo",
		},
		{
			name: "Spanish preference",
			src: &mockSource{
				langs: func() ([]string, error) { return []string{"es_MX"}, nil },
				fetch: func(_ string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Fóó"}]`), nil },
			},
			expected: "Fóó",
		},
		{
			name: "Unsupported preference",
			src: &mockSource{
				langs: func() ([]string, error) { return []string{"de"}, nil },
				fetch: func(_ string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Foo"}]`), nil },
			},
			expected: "Foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set, err := New(test.src)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			set.initWG.Wait()
			T, err := set.Tfunc()
			var initErr string
			if err != nil {
				initErr = err.Error()
			}
			if test.initErr != initErr {
				t.Errorf("Unexpected init error: %s", initErr)
			}
			if err != nil {
				return
			}
			result := T(testID)
			if test.expected != result {
				t.Errorf("Unexpected translation: %s", result)
			}
		})
	}
}

func TestLoadDictionary(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		src      Source
		expected string
		err      string
	}{
		{
			name:   "fetch error",
			locale: "foo",
			src: &mockSource{
				fetch: func(_ string) ([]byte, error) {
					return nil, errors.New("fetch error")
				},
			},
			err: "fetch error",
		},
		{
			name:   "invalid translation data",
			locale: "foo",
			src: &mockSource{
				fetch: func(_ string) ([]byte, error) {
					return []byte("foo"), nil
				},
			},
			err: `no language found in "foo.all.json"`,
		},
		{
			name:   "success",
			locale: "en-us",
			src: &mockSource{
				fetch: func(_ string) ([]byte, error) {
					return []byte(`[{"id":"foo","translation":"Foo"}]`), nil
				},
			},
			expected: "Foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			T, err := loadDictionary(test.locale, test.src)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.err != errMsg {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			result := T(testID)
			if result != test.expected {
				t.Errorf("Unexpected translation: %s", result)
			}
		})
	}
}
