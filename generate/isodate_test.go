package generate_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mwmahlberg/mongen/generate"
	"github.com/stretchr/testify/assert"
)

func (s *GeneratorTestSuite) TestISODate() {
	testCases := []struct {
		desc   string
		start  string
		end    string
		errors bool
	}{
		{
			desc: "Neither start nor end set",
		},
		{
			desc:  "Both 'now'",
			start: "now",
			end:   "now",
		},
		{
			desc:  "Start 'now', end date",
			start: "now",
			end:   time.Now().UTC().Add(time.Hour).Format(generate.MongoDateFormat),
		},
		{
			desc:  "Start 'now', end date",
			start: time.Now().UTC().Add(-1 * time.Hour).Format(generate.MongoDateFormat),
			end:   "now",
		},
		{
			desc:  "Start and End set",
			start: time.Now().UTC().Format(generate.MongoDateFormat),
			end:   time.Now().UTC().Add(time.Hour).Format(generate.MongoDateFormat),
		},
		{
			desc:   "Start > End",
			start:  time.Now().UTC().Add(1 * time.Hour).Format(generate.MongoDateFormat),
			end:    time.Now().UTC().Format(generate.MongoDateFormat),
			errors: true,
		},
		{
			desc:   "Wrong start format",
			start:  time.Now().UTC().Format(time.Kitchen),
			end:    time.Now().UTC().Add(1 * time.Hour).Format(generate.MongoDateFormat),
			errors: true,
		},
		{
			desc:   "Wrong start format",
			start:  time.Now().UTC().Format(generate.MongoDateFormat),
			end:    time.Now().UTC().Add(1 * time.Hour).Format(time.Kitchen),
			errors: true,
		},
	}
	for _, tC := range testCases {
		s.T().Run(tC.desc, func(t *testing.T) {
			// for i := 0; i < 50; i++ {
			d, err := generate.ISODate(tC.start, tC.end)

			if tC.errors {
				assert.Error(t, err)
				assert.Empty(t, d)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, d)

			v := make(map[string]interface{})
			err = json.Unmarshal([]byte(d), &v)
			assert.NoError(t, err, "generator did not return unparseable JSON value '%s': %s", d, err)

			date, err := time.Parse(generate.MongoDateFormat, v["$date"].(string))
			assert.NoErrorf(t, err, "value of '$date' field ('%s')could not be parsed: %s", v["$date"].(string), err)

			if tC.start != "" && tC.start != "now" {
				start, _ := time.Parse(generate.MongoDateFormat, tC.start)
				assert.True(t, date.Equal(start) || date.After(start), "%s is not equal or after %s", date, tC.start)
			}

			if tC.end != "" && tC.end != "now" {
				end, _ := time.Parse(generate.MongoDateFormat, tC.end)
				assert.True(t, date.Before(end) || date.Equal(end))
			}
		})
	}
}

func BenchmarkIsoDate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generate.ISODate("2018-12-20T22:28:00.254Z", "2018-12-20T22:41:05.587Z")
	}
}
