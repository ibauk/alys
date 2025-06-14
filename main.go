package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

const PROGRAMVERSION = "Alys v1.4 Copyright © 2025 Bob Stammers"

// DBNAME names the database file
var DBNAME *string = flag.String("db", "rblr.db", "database file")

// HTTPPort is the web port to serve
var HTTPPort *string = flag.String("port", "80", "Web port")

var EntrantsCSV *string = flag.String("import", "", "CSV to import")

var JGTest *bool = flag.Bool("jg", false, "test jg")

// DBH provides access to the database
var DBH *sql.DB

var STATUSCODES map[string]int

const timefmt = "2006-01-02T15:04"

//go:embed rblr.js
var my_js string

//go:embed rblr.css
var my_css string

var refresher = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>RBLR1000</title>
<style>` + my_css + `</style>
<script>` + my_js + `</script>
</head><body>`

const refreshscript = `<script>setTimeout(function() { window.location=window.location;},15000);</script>`

const timerticker = `var img = document.getElementById('ticker');

var interval = window.setInterval(function(){
    let paused = document.getElementById('timenow');
	if(paused) {paused = paused.getAttribute('data-paused')=='1';}
    if(!paused && img.style.visibility == 'hidden'){
        img.style.visibility = 'visible';
    }else{
        img.style.visibility = 'hidden';
    }
}, 1000);`

var timezone *time.Location

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
	timezone, _ = time.LoadLocation("Europe/London")

}

func beyond24(starttime, finishtime string) bool {

	ok := true
	st, err := time.ParseInLocation(timefmt, starttime, timezone)
	if err != nil {
		ok = false
	}
	ft, err := time.ParseInLocation(timefmt, finishtime, timezone)
	if err != nil {
		ok = false
	}

	hrs := ft.Sub(st).Hours()
	fmt.Printf("%v - %v == %.2f hours\n", finishtime, starttime, hrs)
	return hrs > 24 || !ok
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

	if !checkDB() {
		createDB()
	}

	if *EntrantsCSV != "" {
		LoadEntrantsFromCSV(*EntrantsCSV, "Y", true, false)
		return
	}
	if *JGTest {
		rebuildJGPages()
	}
	http.HandleFunc("/", show_root)
	http.HandleFunc("/getcsv", show_loadCSV)
	http.HandleFunc("/menu", show_menu)
	http.HandleFunc("/about", about_this_program)
	http.HandleFunc("/admin", show_admin)
	http.HandleFunc("/merch", show_shop)
	http.HandleFunc("/stats", show_stats)
	http.HandleFunc("/signin", show_signin)
	http.HandleFunc("/finals", show_finals)
	http.HandleFunc("/edit", edit_entrant)
	http.HandleFunc("/export", export_finishers)
	http.HandleFunc("/checkin", check_in)
	http.HandleFunc("/checkout", check_out)
	http.HandleFunc("/config", show_config)
	http.HandleFunc("/putodo", update_odo)
	http.HandleFunc("/putentrant", update_entrant)
	http.HandleFunc("/upload", load_entrants)
	http.HandleFunc("/just", export_JustGiving)
	http.HandleFunc("/jgtest", doJGTest)
	http.HandleFunc("/search", global_search)
	err = http.ListenAndServe(":"+*HTTPPort, nil)
	if err != nil {
		panic(err)
	}
}

func about_this_program(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<main class="about">`)
	fmt.Fprint(w, `<p class="legal">`+PROGRAMVERSION+"</p>")
	fmt.Fprint(w, "<p>I handle administration for the RBLR1000</p>")
	fmt.Fprint(w, `<p>Importantly, I capture start and finish times and odo readings of entrants. These records form part of the ride proof justifying the issue of IBA certificates.</p>`)
	fp, err := filepath.Abs(*DBNAME)
	checkerr(err)
	fmt.Fprintf(w, `<p>The database is stored in <strong>%v</strong>`, fp)
	hn, err := os.Hostname()
	checkerr(err)
	fmt.Fprintf(w, ` on the server called <strong>%v</strong></p>`, hn)
	fmt.Fprint(w, `<hr><p class="nerdy vspace">This program is written in Go, CSS, HTML and JavaScript and the full source is available at <a href="https://github.com/ibauk/alys" target="github">https://github.com/ibauk/alys</a></p>`)
	fmt.Fprint(w, `</main>`)

	fmt.Fprint(w, `<footer class="about">`)
	fmt.Fprint(w, ` <button class="nav" autofocus onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, "</footer>")

}

func check_in(w http.ResponseWriter, r *http.Request) {
	show_odo(w, r, false, true)
}
func check_out(w http.ResponseWriter, r *http.Request) {
	show_odo(w, r, true, true)
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
	} else if len(res)-dotix == 2 {
		res += "0"
	}
	// 123456.44
	// 012345678
	ix := dotix - 3
	if ix > 1 {
		res = res[0:ix] + "," + res[ix:]
	}
	return res
}

func show_funds_breakdown(w http.ResponseWriter) {

	var sources = map[string]string{
		"EntryDonation": "@ Registration",
		"SquiresCheque": "Cheques",
		"SquiresCash":   "Cash",
		"RBLRAccount":   "&#8658; RBL accounts",
		"JustGivingAmt": "via JustGiving",
	}
	var skeys = []string{"EntryDonation", "SquiresCheque", "SquiresCash", "RBLRAccount", "JustGivingAmt"}

	for _, k := range skeys {
		v := sources[k]
		if k == "JustGivingAmt" {
			n, amt := show_funds_JustGiving()
			if n > 0 {
				fmt.Fprintf(w, `<tr class="subrow"><td>%v (%v)</td><td class="val">&pound;%v</td></tr>`, v, n, format_money(amt))
			}
			continue
		}
		row, err := DBH.Query(fmt.Sprintf("SELECT count(*),ifnull(sum(%v),0) FROM entrants WHERE %v  <>''", k, k))
		checkerr(err)
		defer row.Close()
		var n int64
		var amt string
		if row.Next() {
			err = row.Scan(&n, &amt)
			checkerr(err)
			if n > 0 {
				fmt.Fprintf(w, `<tr class="subrow"><td>%v (%v)</td><td class="val">&pound;%v</td></tr>`, v, n, format_money(amt))

			}
		}
		err = row.Close()
		checkerr(err)
	}

}

func show_funds_JustGiving() (int64, string) {

	row, err := DBH.Query("SELECT ifnull(JustGivingAmt,''),ifnull(JustGivingURL,'') FROM entrants WHERE ifnull(JustGivingAmt,'')<>'' ORDER BY JustGivingURL")
	checkerr(err)
	defer row.Close()
	var amt string
	var lastjgurl string
	var totamt int
	var jgurl string
	var n int64
	for row.Next() {
		err = row.Scan(&amt, &jgurl)
		checkerr(err)
		if len(jgurl) > len(JGV) {
			jgurl = jgurl[len(JGV):]
		}
		p := strings.Index(jgurl, "?")
		if p >= 0 {
			jgurl = jgurl[:p]
		}

		n++
		v := 0
		if jgurl != lastjgurl {
			v = intval(amt)
			totamt += v
		}
		//fmt.Printf("#%v '%v' = %v\n", n, jgurl, v)
		lastjgurl = jgurl
	}
	return n, strconv.Itoa(totamt)
}

func show_root(w http.ResponseWriter, r *http.Request) {

	RallyStatus := getStringFromDB("SELECT RallyStatus FROM config", "S")

	if RallyStatus != "F" {
		show_odo(w, r, true, false)
	} else {
		show_odo(w, r, false, false)
	}

}

func show_shop(w http.ResponseWriter, r *http.Request) {

	var sizes = []string{"S", "M", "L", "XL", "XXL"}

	var ordered = map[string]int{"S": 0, "M": 0, "L": 0, "XL": 0, "XXL": 0}
	var unclaimed = map[string]int{"S": 0, "M": 0, "L": 0, "XL": 0, "XXL": 0}
	var opatches, upatches int

	sqlx := "SELECT ifnull(Tshirt1,''),ifnull(Tshirt2,''),ifnull(Patches,0),EntrantStatus FROM entrants"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	for rows.Next() {
		var t1, t2 string
		var p, es int
		err = rows.Scan(&t1, &t2, &p, &es)
		checkerr(err)
		_, ok := ordered[t1]
		if ok {
			ordered[t1]++
		}

		_, ok = ordered[t2]
		if ok {
			ordered[t2]++
		}
		opatches += p
		if es < STATUSCODES["signedin"] {
			if t1 != "" {
				_, ok := unclaimed[t1]
				if ok {
					unclaimed[t1]++
				} else {
					unclaimed[t1] = 1
				}
			}
			if t2 != "" {
				_, ok := unclaimed[t2]
				if ok {
					unclaimed[t2]++
				} else {
					unclaimed[t2] = 1
				}
			}
			upatches += p

		}
	}

	showzero := r.FormValue("sz") != ""

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<main class="merch">`)

	fmt.Fprint(w, `<h2 class="shop">Merchandise</h2>`)

	fmt.Fprint(w, `<div class="row hdr"><span class="col">Size</span><span class="col">Ordered</span><span class="col">Unclaimed*</span></div>`)
	for _, k := range sizes {
		if unclaimed[k]+ordered[k] > 0 || showzero {
			fmt.Fprintf(w, `<div class="row"><span class="col">%v</span><span class="col">%v</span>`, k, ordered[k])
			fmt.Fprintf(w, `<span class="col">%v</span>`, unclaimed[k])
		}
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprintf(w, `<div class="row"><span class="col">Patches</span><span class="col">%v</span><span class="col">%v</span></div>`, opatches, upatches)

	fmt.Fprint(w, `</main>`)

	fmt.Fprint(w, `<div class="footnote">* Unclaimed means items ordered by entrants who have not yet been signed in.</div>`)

	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
	fmt.Fprint(w, `<footer>`)
	fmt.Fprint(w, `<button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, `</footer>`)

	fmt.Fprint(w, `</body><html>`)
}

func show_stats(w http.ResponseWriter, r *http.Request) {

	fullaccess := r.FormValue("mode") == "f"
	scv := make(map[int]string)
	scv[STATUSCODES["DNS"]] = "not signed in"               // Registered online
	scv[STATUSCODES["confirmedDNS"]] = "withdrawn"          // Confirmed by rider
	scv[STATUSCODES["signedin"]] = "signed in"              // Signed in at Squires
	scv[STATUSCODES["riding"]] = "checked-out (out riding)" // Checked-out at Squires
	scv[STATUSCODES["DNF"]] = "DNF"                         // Ride aborted
	scv[STATUSCODES["finishedOK"]] = "Finished OK"          // Finished inside 24 hours
	scv[STATUSCODES["finished24+"]] = "Finished 24+"        // Finished outside 24 hours

	const showzero = false

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

	fmt.Fprint(w, refresher+refreshscript)

	fmt.Fprint(w, `<main class="stats">`)

	fmt.Fprint(w, `<h2>Live numbers  <span id="ticker">&diams;</span></h2>`)
	fmt.Fprint(w, `<script>`+timerticker+`</script>`)
	fmt.Fprintf(w, `<table><tr><td>registered<br>&nbsp;</td><td class="val">%v<br>&nbsp;</td></tr>`, registered)
	sort.Ints(indexes)

	for _, sc := range indexes {
		if showzero || counts[codedescs[sc]] != 0 {
			fmt.Fprintf(w, `<tr><td>%v</td><td class="val">%v</td></tr>`, scv[sc], counts[codedescs[sc]])
		}
	}
	_, jgamt := show_funds_JustGiving()
	totfundx := getStringFromDB("SELECT SUM(ifnull(EntryDonation,0)+ifnull(SquiresCheque,0)+ifnull(SquiresCash,0)+ifnull(RBLRAccount,0)) AS funds  FROM entrants;", "0.00")
	totfunds := strconv.Itoa(intval(jgamt) + intval(totfundx))
	fmt.Fprintf(w, `<tr><td><br>Funds raised</td><td class="val"><br>&pound;%v</td></tr>`, format_money(totfunds))

	if true {
		show_funds_breakdown(w)
	}

	sqlx := "SELECT count(*) FROM entrants WHERE EntrantStatus IN (" + strconv.Itoa(STATUSCODES["finishedOK"]) + "," + strconv.Itoa(STATUSCODES["finished24+"]) + ") "

	certsWaiting := getIntegerFromDB(sqlx+"AND CertificateAvailable='Y' AND CertificateDelivered<>'Y'", 0)
	certsNeeded := getIntegerFromDB(sqlx+"AND CertificateAvailable<>'Y'", 0)

	if certsWaiting > 0 {
		fmt.Fprintf(w, `<tr><td><br>Finisher certs uncollected</td><td class="val"><br>%v</td></tr>`, certsWaiting)
	}
	if certsNeeded > 0 {
		fmt.Fprintf(w, `<tr><td>Cert reprints needed</td><td class="val">%v</td></tr>`, certsNeeded)
	}

	fmt.Fprint(w, `</table></main>`)
	if fullaccess {
		fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
	}
	fmt.Fprint(w, `<footer>`)

	if fullaccess {
		fmt.Fprint(w, `<button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	} else {
		RallyStatus := getStringFromDB("SELECT RallyStatus FROM config", "S")
		if RallyStatus == "F" {
			fmt.Fprint(w, `<button class="nav" onclick="loadPage('/');">Check-in</button>`)
		} else {
			fmt.Fprint(w, `<button class="nav" onclick="loadPage('/');">Check-out</button>`)
		}
	}
	fmt.Fprint(w, `</footer>`)

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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)
	fmt.Fprint(w, `<main class="config">`)
	fmt.Fprint(w, `<h1>Settings</h1>`)
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
	fmt.Fprint(w, `</main>`)

	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('admin');}}</script>`)

	fmt.Fprint(w, `<footer>`)
	fmt.Fprint(w, ` <button class="nav" onclick="loadPage('admin');">Event setup</button>`)
	fmt.Fprint(w, "</footer>")

}

func show_admin(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, refresher+`<main class="frontmenu">`)
	fmt.Fprint(w, `<h1>RBLR1000 SETUP</h1>`)

	fmt.Fprint(w, `<fieldset class="setup">`)
	fmt.Fprint(w, `<label for="RallyStatus">Currently set for</label> `)
	fmt.Fprint(w, `<select id="RallyStatus" name="RallyStatus" onchange="ocdcfg(this)" data-chg="1">`)
	rs := getStringFromDB("SELECT RallyStatus FROM config", "S")
	sel := ""
	if rs != "F" {
		sel = "selected"
	}
	fmt.Fprintf(w, `<option value="S" %v> Friday / Saturday AM </option>`, sel)
	sel = ""
	if rs == "F" {
		sel = "selected"
	}
	fmt.Fprintf(w, `<option value="F" %v> Saturday / Sunday </option>`, sel)
	fmt.Fprint(w, `</select>`)
	fmt.Fprint(w, `</fieldset>`)
	fmt.Fprint(w, `<button onclick="loadPage('getcsv');" title="Load entrants from CSV prepared by Reglist">Import entrants</button>`)
	fmt.Fprint(w, `<button onclick="this.disabled=true;loadPage('export');" title="Create JSON file for upload to the Rides database">Export results for IBA database</button>`)
	fmt.Fprint(w, `<button onclick="this.disabled=true;loadPage('just');" title="Create CSV file of JustGiving user info">Export JustGiving CSV</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('config');" title="Start times and other variables">Settings</button>`)
	fmt.Fprint(w, `</main>`)
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
	fmt.Fprint(w, `<footer>`)
	fmt.Fprint(w, `<button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, `<button class="nav" onclick="loadPage('about');">About Alys</button>`)
	fmt.Fprint(w, `</footer>`)

	fmt.Fprint(w, `</body><html>`)

}

func show_menu(w http.ResponseWriter, r *http.Request) {

	RallyStatus := getStringFromDB("SELECT RallyStatus FROM config", "S")

	fmt.Fprint(w, refresher+`<main class="frontmenu">`)
	fmt.Fprint(w, `<h1>RBLR1000</h1>`)
	if RallyStatus != "F" {
		fmt.Fprint(w, `<button class="bigscreen" onclick="loadPage('signin');" title="Shows entrants not yet signed-in">SIGN IN(start)</button>`)
		fmt.Fprint(w, `<button onclick="loadPage('checkout');" title="Shows entrants signed-in but not checked-out">CHECK-OUT(start)</button>`)
	} else {
		fmt.Fprint(w, `<button onclick="loadPage('checkin');">CHECK-IN(finish)</button>`)
		fmt.Fprint(w, `<button class="bigscreen" onclick="loadPage('finals');">Verification(finish)</button>`)
	}
	fmt.Fprint(w, `<button onclick="loadPage('stats?mode=f');">Current state of play</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('merch');" title="Summary of pre-ordered T-shirts/patches">Merchandise</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('signin?mode=full');" title="Inspect/update entrant details regardless of status">Full entrant list</button>`)

	fmt.Fprint(w, `<button class="bigscreen" onclick="loadPage('admin');">Event setup</button>`)
	fmt.Fprint(w, `<button onclick="loadPage('search');">Search</button>`)
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
	xtra := ""
	for key, val := range v {
		if key[0:4] != "utm_" { // protect against buffered unencoded urls
			if comma {
				sqlx += ","
			}
			x, err := url.QueryUnescape(val)
			checkerr(err)
			if key == "JustGivingURL" {
				xtra = "JustGivingPSN='" + parseJGPageShortName(x) + "'"
			}
			sqlx += key + "='" + x + "'"
			comma = true
		}
	}
	if xtra != "" {
		sqlx += "," + xtra
	}
	sqlx += " WHERE EntrantID=" + e
	//fmt.Printf("update_entrant: %v\n", sqlx)
	_, err = DBH.Exec(sqlx)
	checkerr(err)
	fmt.Fprint(w, `{"err":false,"msg":"ok"}`)
}

func safesql(x string) string {

	return strings.ReplaceAll(strings.TrimSpace(x), "'", "''")
}
