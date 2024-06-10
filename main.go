package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DBNAME names the database file
var DBNAME *string = flag.String("db", "rblr.db", "database file")

// HTTPPort is the web port to serve
var HTTPPort *string = flag.String("port", "8080", "Web port")

// DBH provides access to the database
var DBH *sql.DB

var STATUSCODES map[string]int

const timefmt = "2006-01-02T15:04"

func init() {
	STATUSCODES = make(map[string]int)
	STATUSCODES["DNS"] = 0          // Registered online
	STATUSCODES["confirmedDNS"] = 1 // Confirmed by rider
	STATUSCODES["signedin"] = 2     // Signed in at Squires
	STATUSCODES["riding"] = 4       // Checked-out at Squires
	STATUSCODES["DNF"] = 6          // Ride aborted
	STATUSCODES["finishedOK"] = 8   // Finished inside 24 hours
	STATUSCODES["finished24+"] = 10 // Finished outside 24 hours

	fmt.Printf("Statuses:\n%v\n\n", STATUSCODES)
}

func getIntegerFromDB(sqlx string, defval int) int {

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		var val int
		rows.Scan(&val)
		return val
	}
	return defval
}

func getStringFromDB(sqlx string, defval string) string {

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		var val string
		rows.Scan(&val)
		return val
	}
	return defval
}

func main() {

	fmt.Println("Hello sailor")
	flag.Parse()

	dbx, _ := filepath.Abs(*DBNAME)
	fmt.Printf("Using %v\n\n", dbx)

	var err error
	DBH, err = sql.Open("sqlite3", dbx)
	if err != nil {
		panic(err)
	}

	sqlx := "SELECT DBInitialised FROM config"
	dbi, _ := strconv.Atoi(getStringFromDB(sqlx, "0"))
	if dbi != 1 {
		fmt.Println("Duff database")
		return
	}

	fmt.Printf("Beyond24? - %v\n", beyond24("", "2024-06-09T19:31"))

	http.HandleFunc("/", central_dispatch)
	http.HandleFunc("/about", about_alys)
	http.HandleFunc("/stats", show_stats)
	http.HandleFunc("/odo", update_odo)
	http.ListenAndServe(":"+*HTTPPort, nil)
}

func about_alys(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello there, I say, I say")
}

func show_stats(w http.ResponseWriter, r *http.Request) {

	const showzero = true
	const refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>Stats</title></head><body>
	<script>setTimeout(function() { window.location=window.location;},15000);</script>`

	registered := getIntegerFromDB("SELECT count(*) FROM entrants", 0)
	codedescs := make(map[int]string)
	counts := make(map[string]int)
	indexes := make([]int, 0)
	for i, v := range STATUSCODES {
		counts[i] = getIntegerFromDB("SELECT count(*) FROM entrants WHERE EntrantStatus="+strconv.Itoa(v), 0)
		codedescs[v] = i
		indexes = append(indexes, v)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)
	fmt.Fprint(w, `<table>`)
	fmt.Fprintf(w, `<tr><td>registered</td><td>%v</td></tr>`, registered)
	sort.Ints(indexes)
	for _, sc := range indexes {
		if showzero || counts[codedescs[sc]] != 0 {
			fmt.Fprintf(w, `<tr><td>%v</td><td>%v</td></tr>`, codedescs[sc], counts[codedescs[sc]])
		}
	}
	fmt.Fprint(w, `</table>`)
	fmt.Fprint(w, `</body><html>`)
}

func storeTimeDB(t time.Time) string {

	res := t.Local().Format(timefmt)
	return res
}

func beyond24(starttime, finishtime string) bool {

	ok := true
	st, err := time.Parse(timefmt, starttime)
	if err != nil {
		ok = false
	}
	ft, err := time.Parse(timefmt, finishtime)
	if err != nil {
		ok = false
	}

	hrs := ft.Sub(st).Hours()
	fmt.Printf("%v - %v == %v hours\n", finishtime, starttime, hrs)
	return hrs > 24 || !ok
}

func update_odo(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("e") == "" || r.FormValue("f") == "" || r.FormValue("v") == "" {
		fmt.Fprint(w, "ok")
		return
	}

	dt := r.FormValue("t")
	if dt == "" {
		dt = storeTimeDB(time.Now())
	}
	sqlx := ""
	switch r.FormValue("f") {
	case "f":
		ns := STATUSCODES["finishedOK"]
		if beyond24(r.FormValue("st"), dt) {
			ns = STATUSCODES["finished24+"]
		}

		sqlx = "OdoFinish=" + r.FormValue("v")
		sqlx += ",FinishTime='" + dt + "'"
		sqlx += ",EntrantStatus=" + strconv.Itoa(ns)
		sqlx += " WHERE EntrantID=" + r.FormValue("e")
		sqlx += " AND FinishTime IS NULL"
		sqlx += " AND EntrantStatus=" + strconv.Itoa(STATUSCODES["riding"])
	case "s":
		sqlx = "OdoStart=" + r.FormValue("v")
		sqlx += ",StartTime='" + dt + "'"
		sqlx += ",EntrantStatus=" + strconv.Itoa(STATUSCODES["riding"])
		sqlx += " WHERE EntrantID=" + r.FormValue("e")
		sqlx += " AND EntrantStatus=" + strconv.Itoa(STATUSCODES["signedin"])
	}
	DBH.Exec("UPDATE entrants SET " + sqlx)

	fmt.Fprint(w, "ok")

}

const mockupFrontPage = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>ALYS</title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body {
	margin: 0;
	font-size: 14pt;
	font-family				: Verdana, Arial, Helvetica, sans-serif; 

}
.topbar {
	background-color: lightgray;
	border none none solid 2px none;
	width: 100%;
	margin: 0;
	padding: 5px;
}
.about {
	float: right;
	padding-right: 1em;
	font-size: 10pt;
	vertical-align: middle;
	display: table-cell;
}
</style>
</head>
<body>
`
const homeIcon = `
<input title="Return to main menu" style="padding:1px;" type="button" value=" 🏠 " onclick="window.location='admin.php'">`

func central_dispatch(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, mockupFrontPage)
	fmt.Fprint(w, `<div class="topbar">`+homeIcon+` 12 Days Euro Rally<span class="about">About ALYS</span></div>`)
}
