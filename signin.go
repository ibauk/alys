package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"text/template"
)

func create_new_entrant() string {

	sqlx := "SELECT max(EntrantID) FROM entrants"

	e := getIntegerFromDB(sqlx, -1) + 1
	if e < 1 {
		return "0"
	}
	res := strconv.Itoa(e)
	sqlx = "INSERT INTO entrants(EntrantID,EntrantStatus) VALUES(" + res + "," + strconv.Itoa(STATUSCODES["signedin"]) + ")"
	_, err := DBH.Exec(sqlx)
	if err != nil {
		return "0"
	}
	return res
}

func edit_entrant(w http.ResponseWriter, r *http.Request) {

	entrant := r.FormValue("e")
	if entrant == "" {
		return
	}

	mode := r.FormValue("m")
	if mode == "" {
		mode = "signin"
	}

	if r.FormValue("e") < "1" {
		entrant = create_new_entrant()
		if entrant < "1" {
			fmt.Fprint(w, `<p>ERROR - can't insert new record!</p>`)
			return
		}
	}
	sqlx := EntrantSQL
	sqlx += " WHERE EntrantID=" + entrant

	rows, err := DBH.Query(sqlx)
	checkerr(err)

	defer rows.Close()

	fmt.Fprint(w, refresher)

	sss, err := template.New("SigninScreenSingle").Parse(SigninScreenSingle)
	checkerr(err)
	fmt.Fprint(w, `<main class="signin">`)
	for rows.Next() {

		var e Entrant

		ScanEntrant(rows, &e)
		e.EditMode = mode
		err = sss.Execute(w, e)
		checkerr(err)

	}

	fmt.Fprint(w, `</main>`)
	fmt.Fprint(w, "<footer>")
	if mode == "signin" {
		fmt.Fprint(w, `<button class="nav" onclick="loadPage('signin');">back to list</button>`)
		fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('signin');}}</script>`)
	}
	fmt.Fprint(w, ` <button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, "</footer>")

}

func show_signin(w http.ResponseWriter, r *http.Request) {

	scv := make(map[int]string)
	scv[STATUSCODES["DNS"]] = "&nbsp;?&nbsp;"                 // Registered online
	scv[STATUSCODES["confirmedDNS"]] = "&nbsp;&#9746;&nbsp;"  // Confirmed by rider
	scv[STATUSCODES["signedin"]] = "&nbsp;&#9745;&nbsp;"      // Signed in at Squires
	scv[STATUSCODES["riding"]] = "&nbsp;&#9745; &#9850;"      // Checked-out at Squires
	scv[STATUSCODES["DNF"]] = "&nbsp;&#9745; &#9760;"         // Ride aborted
	scv[STATUSCODES["finishedOK"]] = "&nbsp;&#9745; &#10004;" // Finished inside 24 hours
	scv[STATUSCODES["finished24+"]] = "&nbsp;&#9745; &check;" // Finished outside 24 hours

	sqlx := EntrantSQL

	mode := r.FormValue("mode")
	if mode == "" {
		mode = "signin"
		scv[STATUSCODES["DNS"]] = "" // This is signin, of course they're not signed in yet
	}
	if mode != "full" {
		sqlx += " WHERE EntrantStatus IN (" + strconv.Itoa(STATUSCODES["DNS"]) + "," + strconv.Itoa(STATUSCODES["confirmedDNS"])
		showSignedin := r.FormValue("all") != ""
		if showSignedin {
			sqlx += "," + strconv.Itoa(STATUSCODES["signedin"])
		}
		sqlx += ")"
	}
	sqlx += " ORDER BY RiderLast,RiderFirst"

	//fmt.Println(sqlx)
	rows, err := DBH.Query(sqlx)
	if err != nil {
		fmt.Println(sqlx)
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	fmt.Fprint(w, refresher)

	action := "Signing-in"
	if mode == "full" {
		action = "Full entrant list"
	}

	fmt.Fprintf(w, `<div class="top"><h2>RBLR1000 - %v  <span id="ticker">&diams;</span></h2></div>`, action)
	fmt.Fprint(w, `<script>`+timerticker+`</script>`)
	fmt.Fprint(w, `<main class="signin">`)

	fmt.Fprint(w, `<div id="signinlist">`)
	n := 0
	oe := true
	for rows.Next() {
		var e Entrant

		ScanEntrant(rows, &e)

		fmt.Fprint(w, `<div class="signinrow `)
		if oe {
			fmt.Fprint(w, "odd")
		} else {
			fmt.Fprint(w, "even")
		}
		oe = !oe
		fmt.Fprintf(w, `" onclick="signin('%v',%v);">`, mode, e.EntrantID)

		val, ok := scv[e.EntrantStatus]
		if !ok {
			val = "!" + strconv.Itoa(e.EntrantStatus)
		}

		fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span> <span class="status">%v</span>`, e.Rider.Last, e.Rider.First, val)

		n++
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div></main>`)
	fmt.Fprint(w, `<footer><button class="nav" onclick="loadPage('menu');">Main menu</button>  `)
	fmt.Fprint(w, ` <input type="checkbox" title="Enable new entrants" onchange="document.getElementById('newentrant').disabled=!this.checked;">`)
	fmt.Fprint(w, ` <button id="newentrant" disabled class="nav" title="Enter unregistered entrant details" onclick="loadPage('edit?e=0');">New Entrant</button>`)
	fmt.Fprint(w, `</footer>`)
	fmt.Fprint(w, `<script>setTimeout(reloadPage,30000);document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)

	fmt.Fprint(w, `</body></html>`)
	//fmt.Printf("Showed %v lines\n", n)
}

func show_finals(w http.ResponseWriter, r *http.Request) {

	//fmt.Printf("show_finals - %v\n", r)
	scv := make(map[int]string)
	scv[STATUSCODES["DNS"]] = "&nbsp;&nbsp;&nbsp;"       // Registered online
	scv[STATUSCODES["confirmedDNS"]] = "DNS"             // Confirmed by rider
	scv[STATUSCODES["signedin"]] = "&nbsp;&#9745;&nbsp;" // Signed in at Squires
	scv[STATUSCODES["riding"]] = "out"                   // Checked-out at Squires
	scv[STATUSCODES["DNF"]] = "dnf"                      // Ride aborted
	scv[STATUSCODES["finishedOK"]] = "fin"               // Finished inside 24 hours
	scv[STATUSCODES["finished24+"]] = "24+"              // Finished outside 24 hours

	MyStatuscodes := []int{
		STATUSCODES["riding"],
		STATUSCODES["DNF"],
		STATUSCODES["finishedOK"],
		STATUSCODES["finished24+"],
	}

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head>
	<meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
	<title>rblr1000</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>`

	sqlx := EntrantSQL
	sqlx += " WHERE EntrantStatus IN (" + strconv.Itoa(STATUSCODES["finishedOK"]) + "," + strconv.Itoa(STATUSCODES["finished24+"])
	sqlx += ")"
	sqlx += " AND CertificateDelivered <> 'Y'"
	sqlx += " ORDER BY RiderLast,RiderFirst"

	//fmt.Println(sqlx)
	rows, err := DBH.Query(sqlx)
	if err != nil {
		fmt.Println(sqlx)
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<div class="top"><h2>RBLR1000 - Verification  <span id="ticker">&diams;</span></h2></div>`)
	fmt.Fprint(w, `<script>`+timerticker+`</script>`)
	fmt.Fprint(w, `<main class="signin">`)

	fmt.Fprint(w, `<div id="signinlist">`)
	n := 0
	oe := true
	itemno := 0
	for rows.Next() {
		var e Entrant

		ScanEntrant(rows, &e)

		itemno++
		fmt.Fprint(w, `<div class="signinrow signout `)
		if oe {
			fmt.Fprint(w, "odd")
		} else {
			fmt.Fprint(w, "even")
		}
		oe = !oe
		fmt.Fprint(w, `">`)

		fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span> `, e.Rider.Last, e.Rider.First)

		fmt.Fprintf(w, `<span class="Route">%v</span> `, e.Route)

		fmt.Fprintf(w, `<span><select id="%ves" name="EntrantStatus" data-e="%v" data-fs="%v" data-dnf="%v" onchange="changeFinalStatus(this);">`, itemno, e.EntrantID, STATUSCODES["finishedOK"], STATUSCODES["DNF"])
		for k, v := range STATUSCODES {
			if slices.Contains(MyStatuscodes, v) {
				fmt.Fprintf(w, `<option value="%v"`, v)
				if e.EntrantStatus == v {
					fmt.Fprint(w, ` selected`)
				}
				fmt.Fprintf(w, `>%v</option>`, k)
			}
		}
		fmt.Fprint(w, `</select></span>`)

		fmt.Fprintf(w, `<span class="field"><select id="%vcs" name="CertificateAD" data-e="%v" data-ca="%v" onchange="changeCertStatus(this);">`, itemno, e.EntrantID, e.CertificateAvailable)

		ca := e.CertificateAvailable != "N"
		cd := e.CertificateDelivered != "N"
		dnf := e.EntrantStatus == STATUSCODES["DNF"]
		fmt.Fprint(w, `<option value="A-D"`)
		if ca && !cd && !dnf {
			fmt.Fprint(w, ` selected`)
		}
		fmt.Fprint(w, `>Certificate available</option>`)

		fmt.Fprint(w, `<option value="A+D"`)
		if ca && cd && !dnf {
			fmt.Fprint(w, ` selected`)
		}
		fmt.Fprint(w, `>Certificate delivered &#10003;</option>`)

		fmt.Fprint(w, `<option value="-A-D"`)
		if !ca && !dnf {
			fmt.Fprint(w, ` selected`)
		}
		fmt.Fprint(w, `>Certificate NEEDED</option>`)

		fmt.Fprint(w, `<option value="dnf"`)
		if dnf {
			fmt.Fprint(w, ` selected`)
		}
		fmt.Fprint(w, `>No certificate due</option>`)

		fmt.Fprint(w, `</select></span>`)

		n++
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div></main>`)
	fmt.Fprint(w, `<footer><button class="nav" onclick="loadPage('menu');">Main menu</button>  `)
	fmt.Fprint(w, `</footer>`)
	fmt.Fprint(w, `<script>setTimeout(reloadPage,30000);setInterval(sendTransactions,1000);document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)

	fmt.Fprint(w, `</body></html>`)
	//fmt.Printf("Showed %v lines\n", n)
}
