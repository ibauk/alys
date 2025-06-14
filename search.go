package main

import (
	"fmt"
	"net/http"
)

func global_search(w http.ResponseWriter, r *http.Request) {

	type table_info struct {
		cid     int
		cname   string
		ctype   string
		notnull int
		defval  any
		pk      int
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<div class="top"><h2>RBLR1000 - Search  </h2></div>`)
	fmt.Fprint(w, `<main class="search">`)

	fmt.Fprint(w, `<form action="/search"`)
	fmt.Fprint(w, `<label for="txt2find">What should I look for?</label> `)
	fmt.Fprintf(w, `<input type="search" autofocus id="txt2find" name="q" value="%v"> `, r.FormValue("q"))
	fmt.Fprint(w, `<button onclick="this.parent.submit()">Find it!</button>`)
	fmt.Fprint(w, `</form>`)

	if r.FormValue("q") == "" {
		fmt.Fprint(w, `</main></div>`)
		fmt.Fprint(w, `<footer><button class="nav" onclick="loadPage('menu');">Main menu</button>  `)
		fmt.Fprint(w, `</footer>`)

		fmt.Fprint(w, `</body></html>`)
		return
	}
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
	sqlx := "SELECT EntrantID,RiderFirst,RiderLast,RiderPhone,Route,ifnull(PillionLast,''),ifnull(PillionFirst,'') FROM entrants WHERE "
	xk := "'%" + safesql(r.FormValue("q")) + "%'"
	wherex := ""
	for n := range cols {
		if wherex != "" {
			wherex += " OR "
		}
		wherex += cols[n] + " LIKE " + xk
	}
	sqlx += wherex
	sqlx += " ORDER BY RiderLast,RiderFirst"
	//fmt.Println(sqlx)
	rows, err = DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	foundit := false
	oe := true
	fmt.Fprint(w, `<div class="results">`)
	for rows.Next() {
		var e, rf, rl, rp, rt, pl, pf string
		err = rows.Scan(&e, &rf, &rl, &rp, &rt, &pl, &pf)
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
		fmt.Fprintf(w, `<span class="route">%v</span>`, rt[2:])
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
