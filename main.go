package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"sort"
	"time"
)

const tsFileName = ".timesheet"

const timeFormat = "15:04"
const dateFormat = "2006-01-02"
const timestampFormat = "2006-01-02 15:04:05"

type entry struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

func (e entry) String() string {
	return fmt.Sprintf("%s | %s", e.Timestamp.Format(timestampFormat), e.Type)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough parameters given")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Couldn't determine current user:", err)
		os.Exit(1)
	}
	tsFile := path.Join(usr.HomeDir, tsFileName)

	entries, err := loadTs(tsFile)
	check(err)

	switch os.Args[1] {
	case "l":
		listAll(entries)
	case "t":
		listToday(entries)
	case "s":
		addStartEntry(entries, tsFile)
	case "e":
		addEndEntry(entries, tsFile)
	case "c":
		calcWorktimeToday(entries)
	case "a":
		calcWorktimeAll(entries)
	}
}

func listAll(entries []entry) {
	for _, e := range entries {
		fmt.Println(e)
	}
}

func listToday(entries []entry) {
	fmt.Println("Today:")
	for _, e := range todaysEntries(entries) {
		fmt.Println(e)
	}
}

func addStartEntry(entries []entry, tsFile string) {
	e := entry{Timestamp: time.Now(), Type: "s"}
	if len(os.Args) > 2 {
		t, err := time.Parse(timeFormat, os.Args[2])
		today := time.Now()
		ts := time.Date(today.Year(), today.Month(), today.Day(), t.Hour(), t.Minute(), t.Second(), 0, today.Location())
		check(err)
		e.Timestamp = ts
		fmt.Println(e.Timestamp.Format(timestampFormat))
	}
	entries = append(entries, e)
	err := saveTs(tsFile, entries)
	check(err)
}

func addEndEntry(entries []entry, tsFile string) {
	e := entry{Timestamp: time.Now(), Type: "e"}
	if len(os.Args) > 2 {
		t, err := time.Parse(timeFormat, os.Args[2])
		today := time.Now()
		ts := time.Date(today.Year(), today.Month(), today.Day(), t.Hour(), t.Minute(), t.Second(), 0, today.Location())
		check(err)
		e.Timestamp = ts
		fmt.Println(e.Timestamp.Format(timestampFormat))
	}
	entries = append(entries, e)
	err := saveTs(tsFile, entries)
	check(err)
}

func calcWorktimeToday(entries []entry) {
	worktime := calcWorktime(todaysEntries(entries), true).Truncate(time.Minute)
	fmt.Println("Working for:", worktime)
	end := time.Now().Add(time.Hour*8 - worktime)
	fmt.Println("Clock off at:", end.Format(timeFormat))
}

func calcWorktimeAll(entries []entry) {
	dayEntries := make(map[string][]entry)
	for _, e := range entries {
		date := e.Timestamp.Format(dateFormat)
		day, ok := dayEntries[date]
		if !ok {
			day = make([]entry, 0, 0)
		}
		day = append(day, e)
		dayEntries[date] = day
	}
	days := make([]string, 0, len(dayEntries))
	for k := range dayEntries {
		days = append(days, k)
	}
	sort.Strings(days)

	var sum time.Duration
	var dayCount int
	for _, d := range days {
		day := dayEntries[d]
		wt := calcWorktime(day, false)
		if wt == 0 {
			fmt.Printf("%s: No end entry\n", d)
		} else {
			fmt.Printf("%s: %v\n", d, wt.Truncate(time.Minute))
			sum += wt
			dayCount++
		}
	}

	expected := (time.Hour * 8) * time.Duration(dayCount)
	diff := sum - expected

	fmt.Printf("Expected: %v\nActual: %v\nDiff: %v\n", expected.Truncate(time.Minute), sum.Truncate(time.Minute), diff.Truncate(time.Minute))
}

func loadTs(filename string) ([]entry, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = ioutil.WriteFile(filename, []byte("[]"), 0600)
		if err != nil {
			return nil, err
		}
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	entries := make([]entry, 0)
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func saveTs(filename string, entries []entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	var indentedData bytes.Buffer

	err = json.Indent(&indentedData, data, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, indentedData.Bytes(), 0600)
}

func todaysEntries(entries []entry) []entry {
	var todaysEntries []entry
	today := time.Now()
	for _, e := range entries {
		if isSameDay(e.Timestamp, today) {
			todaysEntries = append(todaysEntries, e)
		}
	}
	return todaysEntries
}

func calcWorktime(entries []entry, today bool) time.Duration {
	var working bool
	var lastTs time.Time
	var worktime time.Duration

	for _, e := range entries {
		switch e.Type {
		case "s":
			if !working {
				working = true
				lastTs = e.Timestamp
			}
		case "e":
			if !working {
				fmt.Println("Ignoring invalid e entry")
				continue
			}
			working = false
			worktime += e.Timestamp.Sub(lastTs)
		}
	}

	if working {
		if today {
			worktime += time.Now().Sub(lastTs)
		} else {
			worktime = 0
		}
	}

	return worktime
}

func isSameDay(a, b time.Time) bool {
	return a.Day() == b.Day() && a.Month() == b.Month() && a.Year() == b.Year()
}
