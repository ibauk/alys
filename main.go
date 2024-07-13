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
	"text/template"
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
    let paused = document.getElementById('timenow').getAttribute('data-paused')=='1';
    if(!paused && img.style.visibility == 'hidden'){
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
	http.HandleFunc("/admin", show_admin)
	http.HandleFunc("/stats", show_stats)
	http.HandleFunc("/signin", show_signin)
	http.HandleFunc("/edit", edit_entrant)
	http.HandleFunc("/export", export_finishers)
	http.HandleFunc("/checkin", check_in)
	http.HandleFunc("/checkout", check_out)
	http.HandleFunc("/config", show_config)
	http.HandleFunc("/putodo", update_odo)
	http.HandleFunc("/putentrant", update_entrant)
	http.ListenAndServe(":"+*HTTPPort, nil)
}

func about_this_program(w http.ResponseWriter, r *http.Request) {

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>About Alys</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<p class="legal">`+PROGRAMVERSION+"</p>")
	fmt.Fprint(w, "<p>I handle administration for the RBLR1000</p>")
	fp, err := filepath.Abs(*DBNAME)
	checkerr(err)
	fmt.Fprintf(w, `<p>The database is stored in <strong>%v</strong></p>`, fp)
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
	fmt.Fprintf(w, `<table><tr><td>registered<br>&nbsp;</td><td class="val">%v<br>&nbsp;</td></tr>`, registered)
	sort.Ints(indexes)
	for _, sc := range indexes {
		if showzero || counts[codedescs[sc]] != 0 {
			fmt.Fprintf(w, `<tr><td>%v</td><td class="val">%v</td></tr>`, codedescs[sc], counts[codedescs[sc]])
		}
	}
	totfunds := getStringFromDB("SELECT SUM(ifnull(EntryDonation,0)+ifnull(SquiresCheque,0)+ifnull(SquiresCash,0)+ifnull(RBLRAccount,0)+ifnull(JustGivingAmt,0)) AS funds  FROM entrants;", "0.00")
	fmt.Fprintf(w, `<tr><td><br>Funds raised</td><td class="val"><br>&pound;%v</td></tr>`, format_money(totfunds))
	fmt.Fprint(w, `</table></main>`)
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)

	fmt.Fprint(w, `</body><html>`)
}

func storeTimeDB(t time.Time) string {

	res := t.Local().Format(timefmt)
	return res
}

func show_config(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	checkerr(err)

	v := make(map[string]string, 0)
	updt := false
	for key, val := range r.Form {
		v[key] = val[0]
		updt = true
	}

	if updt {
		sqlx := "UPDATE config SET "
		comma := false
		for key, val := range v {
			if comma {
				sqlx += ","
			}
			sqlx += key + "='" + val + "'"
			comma = true
		}
		//fmt.Println(sqlx)
		_, err := DBH.Exec(sqlx)
		checkerr(err)
		fmt.Fprint(w, `{"err":false,"msg":"ok"}`)
		return
	}

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>Config</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	sss, err := template.New("ConfigScreen").Parse(ConfigScreen)
	checkerr(err)

	sqlx := ConfigSQL
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	if rows.Next() {
		var c ConfigRecord
		err = rows.Scan(&c.StartTime, &c.StartCohortMins, &c.ExtraCohorts, &c.RallyStatus)
		checkerr(err)
		err = sss.Execute(w, c)
		checkerr(err)
	}
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

func show_admin(w http.ResponseWriter, r *http.Request) {

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>RBLR1000</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>
	`

	fmt.Fprint(w, refresher+`<main class="frontmenu">`)
	fmt.Fprint(w, `<h1>RBLR1000 ADMINISTRATION</h1>`)
	fmt.Fprint(w, `<button onclick="loadPage('config');">Configuration</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('about');">About Alys</button>`)
	fmt.Fprint(w, `<button onclick="this.disabled=true;loadPage('export');">Export results for IBA database</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, `</main>`)
}

func show_menu(w http.ResponseWriter, r *http.Request) {

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>RBLR1000</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>
	`

	RallyStatus := getStringFromDB("SELECT RallyStatus FROM config", "S")

	fmt.Fprint(w, refresher+`<main class="frontmenu">`)
	fmt.Fprint(w, `<h1>RBLR1000</h1>`)
	if RallyStatus != "F" {
		fmt.Fprint(w, `<button onclick="loadPage('checkout');">CHECK-OUT(start)</button>`)
		fmt.Fprint(w, `<button class="bigscreen" onclick="loadPage('signin');">SIGN IN(start)</button>`)
	} else {
		fmt.Fprint(w, `<button onclick="loadPage('checkin');">CHECK-IN(finish)</button>`)
	}
	fmt.Fprint(w, `<button onclick="loadPage('stats');">show stats</button>`)
	fmt.Fprint(w, `<button class="bigscreen" onclick="loadPage('admin');">administration</button>`)
	fmt.Fprint(w, `</main>`)
}

func update_entrant(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	checkerr(err)
	e := ""
	v := make(map[string]string, 0)
	for key, val := range r.Form {
		if key == "EntrantID" {
			e = val[0]
		} else {
			v[key] = val[0]
		}
	}
	if e == "" {
		fmt.Fprint(w, `{"err": true,"msg":"no entrant"}`)
		return
	}
	if len(v) == 0 {
		fmt.Fprint(w, `{"err":true,"msg":"no data field"}`)
		return
	}
	sqlx := "UPDATE entrants SET "
	comma := false
	for key, val := range v {
		if comma {
			sqlx += ","
		}
		sqlx += key + "='" + val + "'"
		comma = true
	}
	sqlx += " WHERE EntrantID=" + e
	fmt.Println(sqlx)
	_, err = DBH.Exec(sqlx)
	checkerr(err)
	fmt.Fprint(w, `{"err":false,"msg":"ok"}`)
}
