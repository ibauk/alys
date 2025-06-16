package main

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
)

/**
	STATUSCODES = make(map[string]int)
	STATUSCODES["DNS"] = 0          // Registered online
	STATUSCODES["confirmedDNS"] = 1 // Confirmed by rider
	STATUSCODES["signedin"] = 2     // Signed in at Squires
	STATUSCODES["riding"] = 4       // Checked-out at Squires
	STATUSCODES["DNF"] = 6          // Ride aborted
	STATUSCODES["finishedOK"] = 8   // Finished inside 24 hours
	STATUSCODES["finished24+"] = 10 // Finished outside 24 hours
**/

var StatusSearchOptions = map[string]string{
	"not signed-in":             "=0",
	"withdrawn":                 "=1",
	"signed-in,not checked-out": "=2",
	"signed-in":                 ">=2",
	"checked-out":               ">=4",
	"DNF":                       "=6",
	"Finishers (incl Late)":     ">=8",
	"Still out riding":          "=4",
	"Late Finishers":            "=10",
	"Unverified Finishers":      ">=8 AND Verified<>'Y'",
}

func global_search(w http.ResponseWriter, r *http.Request) {

	type table_info struct {
		cid     int
		cname   string
		ctype   string
		notnull int
		defval  any
		pk      int
	}

	scv := status_icon_map()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<div class="top"><h2>RBLR1000 - Search  </h2></div>`)
	fmt.Fprint(w, `<main class="search">`)

	fmt.Fprint(w, `<form action="/search">`)
	//	fmt.Fprint(w, `<label for="txt2find">What should I look for?</label> `)
	fmt.Fprintf(w, `<input type="search" placeholder="What should I look for?" autofocus id="txt2find" name="q" value="%v"> `, r.FormValue("q"))

	fmt.Fprint(w, `<select name="qr"> `)

	fmt.Fprint(w, `<option value="" >any route</option>`)
	rso := slices.Sorted(maps.Keys(RouteSearchOptions))
	for _, rl := range rso {
		fmt.Fprintf(w, `<option value="%v" `, rl)
		if rl == r.FormValue("qr") {
			fmt.Fprint(w, `selected`)
		}
		fmt.Fprintf(w, `>%v</option>`, rl)
	}
	fmt.Fprint(w, `</select> `)

	fmt.Fprint(w, ` <select name="qs"> `)
	fmt.Fprint(w, `<option value="">any status</option>`)
	sso := slices.Sorted(maps.Keys(StatusSearchOptions))
	for _, so := range sso {
		fmt.Fprintf(w, `<option value="%v" `, so)
		if r.FormValue("qs") == so {
			fmt.Fprint(w, ` selected `)
		}
		fmt.Fprintf(w, `>%v</option>`, so)
	}
	fmt.Fprint(w, `</select> `)

	fmt.Fprint(w, `<button onclick="this.parent.submit()">Find it!</button>`)
	fmt.Fprint(w, `</form>`)

	rows, err := DBH.Query("pragma table_info(entrants)")
	checkerr(err)
	defer rows.Close()
	cols := make([]string, 0)
	for rows.Next() {
		var ti table_info
		err = rows.Scan(&ti.cid, &ti.cname, &ti.ctype, &ti.notnull, &ti.defval, &ti.pk)
		checkerr(err)
		cols = append(cols, ti.cname)
	}
	rows.Close()
	sqlx := "SELECT EntrantID,RiderFirst,RiderLast,RiderPhone,Route,ifnull(PillionLast,''),ifnull(PillionFirst,''),EntrantStatus FROM entrants "
	xk := "'%" + safesql(r.FormValue("q")) + "%'"
	wherex := ""
	for n := range cols {
		if wherex != "" {
			wherex += " OR "
		}
		wherex += cols[n] + " LIKE " + xk
	}
	if r.FormValue("q") != "" {
		sqlx += " WHERE (" + wherex + ")"
	}
	if r.FormValue("qr") != "" {
		if r.FormValue("q") != "" {
			sqlx += " AND "
		} else {
			sqlx += " WHERE "
		}
		sqlx += " Route " + RouteSearchOptions[r.FormValue("qr")]
	}

	if r.FormValue("qs") != "" {
		if r.FormValue("q") != "" || r.FormValue("qr") != "" {
			sqlx += " AND "
		} else {
			sqlx += " WHERE "
		}
		sqlx += " EntrantStatus " + StatusSearchOptions[r.FormValue("qs")]
	}

	sqlx += " ORDER BY RiderLast,RiderFirst"
	if r.FormValue("debug") == "1" {
		fmt.Println(sqlx)
	}
	rows, err = DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	foundit := false
	oe := true
	fmt.Fprint(w, `<div class="results">`)
	for rows.Next() {
		var e, rf, rl, rp, rt, pl, pf string
		var es int
		err = rows.Scan(&e, &rf, &rl, &rp, &rt, &pl, &pf, &es)
		checkerr(err)
		fmt.Fprint(w, `<div class="resultline`)
		if oe {
			fmt.Fprint(w, ` odd`)
		} else {
			fmt.Fprint(w, ` even`)
		}
		oe = !oe
		fmt.Fprintf(w, `" onclick="signin('full','%v');">`, e)
		fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span>`, rl, rf)
		fmt.Fprintf(w, `<span class="phone">%v</span>`, rp)
		fmt.Fprintf(w, `<span class="route">%v</span>`, DisplayRoute(rt))
		fmt.Fprintf(w, `<span class="status">%v</span>`, scv[es])
		if pl != "" {
			fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span>`, pl, pf)
		}
		fmt.Fprint(w, `</div>`)
		foundit = true
	}
	if !foundit {
		fmt.Fprint(w, `<p>Sorry, nothing found &#9785;</p>`)
	}
	fmt.Fprint(w, `</div></main></div>`)

	fmt.Fprint(w, `<footer><button class="nav" onclick="loadPage('menu');">Main menu</button>  `)
	fmt.Fprint(w, `</footer>`)

	fmt.Fprint(w, `</body></html>`)

}
