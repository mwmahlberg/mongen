package generate

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// MongoDateFormat differs from time.RFC3339Nano as it only
// has millisecond precision.
const MongoDateFormat = "2006-01-02T15:04:05.999+07:00"

// ISODate returns a string representing an ISODate as used by MongoDB.
// However, min and max can be in either MongoDateFormat or time.RFC3339Nano
// as the latter is a superset of the former.
// If min is empty, it is set to "1677-09-21T00:12:43.145Z".
// If max is empty, it is set to "2262-04-11T23:47:16.854Z".
// Both min and max can be replaced with the keyword "now", which sets either of the values
// to the current date and time. Setting both to now returns the current time in UTC.
func ISODate(min, max string) (string, error) {
	var minDate, maxDate time.Time
	var err error

	if min == "now" && max == "now" {
		return ISODateNow(), nil
	}

	switch min {
	case "":
		minDate = time.Unix(0, -math.MaxInt64-1).UTC()
	case "now":
		minDate = time.Now().UTC()
	default:
		if minDate, err = time.Parse(MongoDateFormat, min); err != nil {
			return "", fmt.Errorf("Error parsing '%s': %s", min, err)
		}
	}

	switch max {
	case "":
		maxDate = time.Unix(0, 1e9).UTC()
	case "now":
		maxDate = time.Now().UTC()
	default:
		if maxDate, err = time.Parse(MongoDateFormat, max); err != nil {
			return "", fmt.Errorf("Error parsing '%s': %s", max, err)
		}
	}

	if minDate.After(maxDate) {
		return "", fmt.Errorf("MinDate is after MaxDate")
	}

	val := maxDate.UnixNano() - minDate.UnixNano()
	gen := rand.Int63n(int64(math.Abs(float64(val)))) + minDate.UnixNano()

	return fmt.Sprintf(`{"$date":"%s"}`, time.Unix(0, gen).UTC().Format(MongoDateFormat)), nil
	// return "ISODate(\"" + time.Unix(0, gen).UTC().Format(MongoDateFormat) + "\")", nil

}

// ISODateNow returns the current time in the UTC timezone formatted to MongoDateFormat.
func ISODateNow() string {
	return fmt.Sprintf(`{"$date":"%s"}`, time.Now().UTC().Format(MongoDateFormat))
	// return "ISODate(\"" + time.Now().UTC().Format(MongoDateFormat) + "\")"
}
