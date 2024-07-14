package main

import (
	"fmt"
	"net/http"
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
	sqlx = "INSERT INTO entrants(EntrantID) VALUES(" + res + ")"
	_, err := DBH.Exec(sqlx)
	if err != nil {
		return "0"
	}
	return res
}

func edit_entrant(w http.ResponseWriter, r *http.Request) {

	var refresher = `<!DOCTYPE html>
		<html lang="en">
		<head><title>Signin</title>
		<style>` + my_css + `</style>
		<script>` + my_js + `</script>
		</head><body>`

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
	for rows.Next() {

		var e Entrant

		ScanEntrant(rows, &e)
		e.EditMode = mode
		err = sss.Execute(w, e)
		checkerr(err)

	}

	fmt.Fprint(w, "<nav>")
	if mode == "signin" {
		fmt.Fprint(w, `<button class="nav" onclick="loadPage('signin');">back to list</button>`)
		fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('signin');}}</script>`)
	}
	fmt.Fprint(w, ` <button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, "</nav")

}

func show_signin(w http.ResponseWriter, r *http.Request) {

	scv := make(map[int]string)
	scv[STATUSCODES["DNS"]] = "&nbsp;&nbsp;&nbsp;"       // Registered online
	scv[STATUSCODES["confirmedDNS"]] = "DNS"             // Confirmed by rider
	scv[STATUSCODES["signedin"]] = "&nbsp;&#9745;&nbsp;" // Signed in at Squires
	scv[STATUSCODES["riding"]] = "out"                   // Checked-out at Squires
	scv[STATUSCODES["DNF"]] = "dnf"                      // Ride aborted
	scv[STATUSCODES["finishedOK"]] = "fin"               // Finished inside 24 hours
	scv[STATUSCODES["finished24+"]] = "24+"              // Finished outside 24 hours

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>Signin</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>`

	sqlx := EntrantSQL
	sqlx += " WHERE EntrantStatus IN (" + strconv.Itoa(STATUSCODES["DNS"]) + "," + strconv.Itoa(STATUSCODES["confirmedDNS"])
	showSignedin := r.FormValue("all") != ""
	if showSignedin {
		sqlx += "," + strconv.Itoa(STATUSCODES["signedin"])
	}
	sqlx += ")"
	sqlx += " ORDER BY RiderLast,RiderFirst"

	//fmt.Println(sqlx)
	rows, err := DBH.Query(sqlx)
	if err != nil {
		fmt.Println(sqlx)
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<div class="top"><h2>RBLR1000 - Signing-in</h2></div>`)
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
		fmt.Fprintf(w, `" onclick="signin(%v);">`, e.EntrantID)

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
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)

	fmt.Fprint(w, `</body></html>`)
	//fmt.Printf("Showed %v lines\n", n)
}

func show_finals(w http.ResponseWriter, r *http.Request) {

	scv := make(map[int]string)
	scv[STATUSCODES["DNS"]] = "&nbsp;&nbsp;&nbsp;"       // Registered online
	scv[STATUSCODES["confirmedDNS"]] = "DNS"             // Confirmed by rider
	scv[STATUSCODES["signedin"]] = "&nbsp;&#9745;&nbsp;" // Signed in at Squires
	scv[STATUSCODES["riding"]] = "out"                   // Checked-out at Squires
	scv[STATUSCODES["DNF"]] = "dnf"                      // Ride aborted
	scv[STATUSCODES["finishedOK"]] = "fin"               // Finished inside 24 hours
	scv[STATUSCODES["finished24+"]] = "24+"              // Finished outside 24 hours

	var refresher = `<!DOCTYPE html>
	<html lang="en">
	<head><title>Final checks</title>
	<style>` + my_css + `</style>
	<script>` + my_js + `</script>
	</head><body>`

	sqlx := EntrantSQL
	sqlx += " WHERE EntrantStatus IN (" + strconv.Itoa(STATUSCODES["DNF"]) + "," + strconv.Itoa(STATUSCODES["finishedOK"]) + "," + strconv.Itoa(STATUSCODES["finished24+"])
	sqlx += ")"
	sqlx += " ORDER BY RiderLast,RiderFirst"

	//fmt.Println(sqlx)
	rows, err := DBH.Query(sqlx)
	if err != nil {
		fmt.Println(sqlx)
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<div class="top"><h2>RBLR1000 - Verification</h2></div>`)
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
		fmt.Fprintf(w, `" onclick="signin(%v);">`, e.EntrantID)

		val, ok := scv[e.EntrantStatus]
		if !ok {
			val = "!" + strconv.Itoa(e.EntrantStatus)
		}

		fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span> <span class="status">%v</span>`, e.Rider.Last, e.Rider.First, val)

		n++
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div>`)
	fmt.Fprint(w, `<hr><button class="nav" onclick="loadPage('menu');">Main menu</button>  `)
	fmt.Fprint(w, ` <input type="checkbox" title="Enable new entrants" onchange="document.getElementById('newentrant').disabled=!this.checked;">`)
	fmt.Fprint(w, ` <button id="newentrant" disabled class="nav" title="Enter unregistered entrant details" onclick="loadPage('edit?e=0');">New Entrant</button>`)
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)

	fmt.Fprint(w, `</main></body></html>`)
	//fmt.Printf("Showed %v lines\n", n)
}
