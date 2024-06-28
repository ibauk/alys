package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

const PROGRAMVERSION = "Alys v0.1  Copyright (c) 2024 Bob Stammers"

// DBNAME names the database file
var DBNAME *string = flag.String("db", "rblr.db", "database file")

// HTTPPort is the web port to serve
var HTTPPort *string = flag.String("port", "80", "Web port")

// DBH provides access to the database
var DBH *sql.DB

var STATUSCODES map[string]int

const timefmt = "2006-01-02T15:04"

//go:embed rblr.js
var my_js string

//go:embed rblr.css
var my_css string

const timerticker = `var img = document.getElementById('ticker');

var interval = window.setInterval(function(){
    if(img.style.visibility == 'hidden'){
        img.style.visibility = 'visible';
    }else{
        img.style.visibility = 'hidden';
    }
}, 1000);`

func init() {
	STATUSCODES = make(map[string]int)
	STATUSCODES["DNS"] = 0          // Registered online
	STATUSCODES["confirmedDNS"] = 1 // Confirmed by rider
	STATUSCODES["signedin"] = 2     // Signed in at Squires
	STATUSCODES["riding"] = 4       // Checked-out at Squires
	STATUSCODES["DNF"] = 6          // Ride aborted
	STATUSCODES["finishedOK"] = 8   // Finished inside 24 hours
	STATUSCODES["finished24+"] = 10 // Finished outside 24 hours

	//fmt.Printf("Statuses:\n%v\n\n", STATUSCODES)
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

	fmt.Println(PROGRAMVERSION)
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

	//	fmt.Printf("Beyond24? - %v\n", beyond24("", "2024-06-09T19:31"))

	http.HandleFunc("/", show_menu)
	http.HandleFunc("/menu", show_menu)
	http.HandleFunc("/about", about_this_program)
	http.HandleFunc("/stats", show_stats)
	http.HandleFunc("/signin", show_signin)
	http.HandleFunc("/edit", edit_entrant)
	http.HandleFunc("/checkin", check_in)
	http.HandleFunc("/checkout", check_out)
	http.HandleFunc("/putodo", update_odo)
	http.ListenAndServe(":"+*HTTPPort, nil)
}

func about_this_program(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello there, I say, I say")
}

func check_in(w http.ResponseWriter, r *http.Request) {
	show_odo(w, r, false)
}
func check_out(w http.ResponseWriter, r *http.Request) {
	show_odo(w, r, true)
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func format_money(moneyamt string) string {

	res := moneyamt
	dotix := strings.Index(res, ".")
	if dotix < 0 {
		res += ".00"
	}
	// 123456.44
	// 012345678
	ix := dotix - 3
	if ix > 1 {
		res = res[0:ix] + "," + res[ix:]
	}
	return res
}

func show_stats(w http.ResponseWriter, r *http.Request) {

	const showzero = false
	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>Stats</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>
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
	fmt.Fprint(w, `<main class="stats">`)
	fmt.Fprint(w, `<button class="nav" onclick="loadPage('menu');">Main menu</button>`)

	fmt.Fprint(w, `<h2>Live numbers  <span id="ticker">&diams;</span></h2>`)
	fmt.Fprint(w, `<script>`+timerticker+`</script>`)
	fmt.Fprintf(w, `<table><tr><td>registered<br></td><td class="val">%v<br></td></tr>`, registered)
	sort.Ints(indexes)
	for _, sc := range indexes {
		if showzero || counts[codedescs[sc]] != 0 {
			fmt.Fprintf(w, `<tr><td>%v</td><td class="val">%v</td></tr>`, codedescs[sc], counts[codedescs[sc]])
		}
	}
	totfunds := getStringFromDB("SELECT SUM(ifnull(EntryDonation,0)+ifnull(SquiresCheque,0)+ifnull(SquiresCash,0)+ifnull(RBLRAccount,0)+ifnull(JustGivingAmt,0)) AS funds  FROM entrants;", "0.00")
	fmt.Fprintf(w, `<tr><td><br>Funds raised</td><td class="val"><br>&pound;%v</td></tr>`, format_money(totfunds))
	fmt.Fprint(w, `</table></main>`)

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

// When showing the odo capture list, this sets the time shown in the header
func get_odolist_start_time(ischeckout bool) (string, int, int) {

	res := storeTimeDB((time.Now()))
	if !ischeckout {
		return res, 0, 0
	}
	// Need to show next available start rather than real time
	st := getStringFromDB("SELECT StartTime FROM config", "")
	if st == "" {
		return res, 0, 0
	}
	res = res[0:11] + st
	mins := getIntegerFromDB("SELECT StartCohortMins FROM config", 10)
	xtra := getIntegerFromDB("SELECT ExtraCohorts FROM config", 3)
	return res, mins, xtra

}

func show_menu(w http.ResponseWriter, r *http.Request) {

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>RBLR1000</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>
	`

	fmt.Fprint(w, refresher+`<main class="frontmenu">`)
	fmt.Fprint(w, `<h1>RBLR1000</h1>`)
	fmt.Fprint(w, `<button onclick="loadPage('checkout');">CHECK-OUT(start)</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('checkin');">CHECK-IN(finish)</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('stats');">show stats</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('signin');">SIGN IN(start)</button>`)
	fmt.Fprint(w, `<button>administration</button>`)
	fmt.Fprint(w, `</main>`)
}

func show_odo(w http.ResponseWriter, r *http.Request, showstart bool) {

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>Odo capture</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>`

	sqlx := "SELECT EntrantID,RiderFirst,RiderLast,ifnull(OdoStart,''),ifnull(StartTime,''),ifnull(OdoFinish,''),ifnull(FinishTime,''),EntrantStatus,OdoKms"
	sqlx += " FROM entrants WHERE "
	st, gap, xtra := get_odolist_start_time(showstart)
	sclist := ""
	if showstart {
		sclist = strconv.Itoa(STATUSCODES["signedin"])
	} else {
		sclist = strconv.Itoa(STATUSCODES["riding"]) + "," + strconv.Itoa(STATUSCODES["DNF"])
	}
	sqlx += " EntrantStatus IN (" + sclist + ")"
	sqlx += " ORDER BY RiderLast,RiderFirst"
	//fmt.Println(sqlx)
	rows, _ := DBH.Query(sqlx)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<div id="odohdr">`)

	odoname := ""
	if showstart {
		fmt.Fprint(w, " START")
		odoname = "s"
	} else {
		fmt.Fprint(w, " FINISH")
		odoname = "f"
	}
	fmt.Fprintf(w, ` <span id="timenow" data-time="%v" data-refresh="1000" data-pause="120000" data-paused="0"`, st)
	if showstart {
		fmt.Fprintf(w, ` data-gap="%v" data-xtra="%v"`, gap, xtra)
	}
	fmt.Fprintf(w, ` onclick="clickTime();">%v</span>`, st[11:16])

	fmt.Fprint(w, ` <span id="ticker">&diams;</span>`)
	fmt.Fprint(w, `<script>`+timerticker+`</script>`)

	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<script>refreshTime(); timertick = setInterval(refreshTime,1000);</script>`)

	fmt.Fprint(w, `<div id="odolist">`)
	oe := true
	itemno := 0
	for rows.Next() {
		var EntrantID int
		var RiderFirst, RiderLast, OdoStart, StartTime, OdoFinish, FinishTime string
		var EntrantStatus int
		var OdoKms int
		rows.Scan(&EntrantID, &RiderFirst, &RiderLast, &OdoStart, &StartTime, &OdoFinish, &FinishTime, &EntrantStatus, &OdoKms)
		itemno++
		fmt.Fprint(w, `<div class="odorow `)
		if oe {
			fmt.Fprint(w, "odd")
		} else {
			fmt.Fprint(w, "even")
		}
		oe = !oe
		fmt.Fprint(w, `">`)

		fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span> `, RiderLast, RiderFirst)
		pch := "finish odo"
		val := OdoFinish
		if showstart {
			pch = "start odo"
			val = OdoStart
		}
		fmt.Fprintf(w, `<span><input id="%v" data-e="%v" name="%v" type="number" class="bignumber" oninput="oi(this);" onchange="oc(this);" min="0" placeholder="%v" value="%v"></span>`, itemno, EntrantID, odoname, pch, val)
		fmt.Fprint(w, `</div>`)

	}
	fmt.Fprint(w, `</div><hr><button class="nav" onclick="loadPage('menu');">Main menu</button></body></html>`)
}

func update_odo(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Here we go")
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
		sqlx += " AND EntrantStatus IN (" + strconv.Itoa(STATUSCODES["riding"]) + strconv.Itoa(STATUSCODES["DNF"]) + ")"
	case "s":
		sqlx = "OdoStart=" + r.FormValue("v")
		sqlx += ",StartTime='" + dt + "'"
		sqlx += ",EntrantStatus=" + strconv.Itoa(STATUSCODES["riding"])
		sqlx += " WHERE EntrantID=" + r.FormValue("e")
		sqlx += " AND EntrantStatus IN (" + strconv.Itoa(STATUSCODES["signedin"]) + "," + strconv.Itoa(STATUSCODES["riding"]) + ")"
	}
	fmt.Println(sqlx)
	DBH.Exec("UPDATE entrants SET " + sqlx)

	fmt.Fprint(w, "ok")

}
