package timeofday

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TimeOfDay struct {
	hour   int
	minute int
	loc    *time.Location
}

func New(hour, minute int, loc *time.Location) (*TimeOfDay, error) {
	if hour < 0 || hour > 24 {
		return nil, fmt.Errorf("Hour integer must be between 0 and 24, inclusive")
	}
	hour %= 24 //24 = 0
	if minute < 0 || minute > 59 {
		return nil, fmt.Errorf("Minute integer must be between 0 and 59, inclusive")
	}

	if loc == nil {
		loc = time.Local
	}

	return &TimeOfDay{
		hour:   hour,
		minute: minute,
		loc:    loc,
	}, nil
}

func NewFromString(timeSpec string, loc *time.Location) (*TimeOfDay, error) {
	tokens, err := tokenizeTimeSpec(timeSpec)
	if err != nil {
		return nil, err
	}

	return tokens.Parse(loc)
}

type timeSpecTokens struct {
	Hour     string
	Minute   string
	Meridiem string
}

func tokenizeTimeSpec(spec string) (*timeSpecTokens, error) {
	var (
		runes  = []rune(strings.TrimSpace(strings.ToUpper(spec)))
		state  int
		pos    int
		err    error
		tokens = &timeSpecTokens{}
	)
	wrapErrPosition := func(err error, position int) error {
		return fmt.Errorf("Parse error at position %d (zero-indexed): %s", position, err)
	}

	for {
		if pos >= len(runes) {
			break
		}

		c := runes[pos]
		//fmt.Printf("c: `%c'\tstate: `%d'\tpos: `%d'\n", c, state, pos)
		switch state {
		case 0: //hour
			if err = validRune(c, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'); err != nil {
				return nil, wrapErrPosition(err, pos)
			}

			tokens.Hour = string(c)
		case 1: //second hour character? or colon. or start of meridiem
			if err = validRune(c,
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ' ', 'A', 'P',
			); err != nil {
				return nil, wrapErrPosition(err, pos)
			}
			if c == ':' || c == ' ' || c == 'A' || c == 'P' {
				state = 2
				continue
			}
			tokens.Hour += string(c)

		case 2: //colon or start of meridiem (post-hour token)
			if err = validRune(c, ':', ' ', 'A', 'P'); err != nil {
				return nil, wrapErrPosition(err, pos)
			}

			if c != ':' { //must be meridiem
				state = 5
				continue
			}

		case 3: //minute (first character)
			if err = validRune(c, '0', '1', '2', '3', '4', '5'); err != nil {
				return nil, wrapErrPosition(err, pos)
			}

			tokens.Minute = string(c)

		case 4: //minute (second character)
			if err = validRune(c, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'); err != nil {
				return nil, wrapErrPosition(err, pos)
			}

			tokens.Minute += string(c)

		case 5: //meridiem 1
			if err = validRune(c, ' ', 'A', 'P'); err != nil {
				return nil, wrapErrPosition(err, pos)
			}

			if c == ' ' {
				pos++
				continue
			}

			tokens.Meridiem += string(c)

		case 6: //meridiem 2
			if err = validRune(c, 'M'); err != nil {
				return nil, wrapErrPosition(err, pos)
			}

			tokens.Meridiem += "M"

		case 7: //too much...
			return nil, fmt.Errorf("Extra character `%c': Expected <end>", c)
		}
		pos++
		state++
	}

	const eofError = "Unexpected end of input"
	switch state {
	case 0, 3, 4:
		return nil, fmt.Errorf("%s: Expected number", eofError)
	case 6:
		return nil, fmt.Errorf("%s: Expected 'M'", eofError)
	}

	return tokens, err
}

func validRune(c rune, valid ...rune) error {
	for _, r := range valid {
		if c == r {
			return nil
		}
	}
	validStrings := make([]string, len(valid))
	for i := range valid {
		validStrings[i] = string(valid[i])
	}
	return fmt.Errorf("Invalid rune `%c': Expected one of `%s'",
		c, strings.Join(validStrings, "', `"))
}

func (tokens *timeSpecTokens) Parse(loc *time.Location) (*TimeOfDay, error) {
	hour, err := strconv.Atoi(tokens.Hour)
	if err != nil {
		return nil, fmt.Errorf("Could not parse hour as integer: %s", err)
	}

	minute, err := strconv.Atoi(tokens.Minute)
	if err != nil {
		return nil, fmt.Errorf("Could not parse minute as integer: %s", err)
	}

	if tokens.Meridiem != "" && hour == 0 {
		return nil, fmt.Errorf("Cannot have meridiem and hour value of 0")
	}

	switch tokens.Meridiem {
	case "":
	case "AM":
		if hour == 12 {
			hour = 0
		}
	case "PM":
		if hour > 12 {
			return nil, fmt.Errorf("Cannot have hour greater than 12 with meridiem PM")
		} else if hour != 12 {
			hour += 12
		}
	default:
		return nil, fmt.Errorf("Unknown meridiem spec `%s'", tokens.Meridiem)
	}

	return New(hour, minute, loc)
}

func (t *TimeOfDay) Next() time.Time {
	return t.NextAfter(time.Now())
}

func (t *TimeOfDay) NextAfter(after time.Time) time.Time {
	ret := time.Date(after.Year(), after.Month(), after.Day(), t.hour, t.minute, 0, 0, t.loc)
	if ret.Before(after) {
		ret = time.Date(after.Year(), after.Month(), after.Day()+1, t.hour, t.minute, 0, 0, t.loc)
	}

	return ret
}

func (t *TimeOfDay) Hour() int {
	return time.Date(2006, time.January, 2, t.hour, t.minute, 0, 0, t.loc).Hour()
}

func (t *TimeOfDay) Minute() int {
	return time.Date(2006, time.January, 2, t.hour, t.minute, 0, 0, t.loc).Minute()
}

func (t *TimeOfDay) Location() *time.Location {
	return t.loc
}
