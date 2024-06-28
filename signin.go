package main

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"
)

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

	sqlx := "SELECT EntrantID,RiderFirst,RiderLast,ifnull(RiderIBA,''),ifnull(RiderRBLR,''),ifnull(Email,''),ifnull(Phone,''),ifnull(RiderAddress,'')"
	sqlx += ",ifnull(PillionFirst,''),ifnull(PillionLast,''),ifnull(PillionIBA,''),ifnull(PillionRBLR,''),ifnull(PillionEmail,''),ifnull(PillionPhone,''),ifnull(PillionAddress,'')"
	sqlx += ",ifnull(Bike,'motorbike'),ifnull(BikeReg,'')"
	sqlx += ",ifnull(NokName,''),ifnull(NokRelation,''),ifnull(NokPhone,'')"
	sqlx += ",ifnull(OdoStart,''),ifnull(StartTime,''),ifnull(OdoFinish,''),ifnull(FinishTime,''),EntrantStatus,OdoKms,ifnull(Route,'')"
	sqlx += ",ifnull(EntryDonation,''),ifnull(SquiresCash,''),ifnull(SquiresCheque,''),ifnull(RBLRAccount,''),ifnull(JustGivingAmt,'')"
	sqlx += " FROM entrants"
	sqlx += " WHERE EntrantID=" + entrant

	rows, err := DBH.Query(sqlx)
	checkerr(err)

	fmt.Fprint(w, refresher)
	sss, err := template.New("SigninScreenSingle").Parse(SigninScreenSingle)
	checkerr(err)
	for rows.Next() {

		var e Entrant

		err := rows.Scan(&e.EntrantID, &e.Rider.First, &e.Rider.Last, &e.Rider.IBA, &e.Rider.RBLR, &e.Rider.Email, &e.Rider.Phone, &e.Rider.Address, &e.Pillion.First, &e.Pillion.Last, &e.Pillion.IBA, &e.Pillion.RBLR, &e.Pillion.Email, &e.Pillion.Phone, &e.Pillion.Address, &e.Bike, &e.BikeReg, &e.NokName, &e.NokRelation, &e.NokPhone, &e.OdoStart, &e.StartTime, &e.OdoFinish, &e.FinishTime, &e.EntrantStatus, &e.OdoKms, &e.Route, &e.FundsRaised.EntryDonation, &e.FundsRaised.SquiresCash, &e.FundsRaised.SquiresCheque, &e.FundsRaised.RBLRAccount, &e.FundsRaised.JustGivingAmt)
		checkerr(err)
		err = sss.Execute(w, e)
		checkerr(err)

	}
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

	sqlx := "SELECT EntrantID,RiderFirst,RiderLast,ifnull(RiderIBA,''),ifnull(RiderRBLR,''),ifnull(Email,''),ifnull(Phone,''),ifnull(RiderAddress,'')"
	sqlx += ",ifnull(PillionFirst,''),ifnull(PillionLast,''),ifnull(PillionIBA,''),ifnull(PillionRBLR,''),ifnull(PillionEmail,''),ifnull(PillionPhone,''),ifnull(PillionAddress,'')"
	sqlx += ",ifnull(Bike,'motorbike'),ifnull(BikeReg,'')"
	sqlx += ",ifnull(NokName,''),ifnull(NokRelation,''),ifnull(NokPhone,'')"
	sqlx += ",ifnull(OdoStart,''),ifnull(StartTime,''),ifnull(OdoFinish,''),ifnull(FinishTime,''),EntrantStatus,OdoKms,ifnull(Route,'')"
	sqlx += ",ifnull(EntryDonation,''),ifnull(SquiresCash,''),ifnull(SquiresCheque,''),ifnull(RBLRAccount,''),ifnull(JustGivingAmt,'')"
	sqlx += " FROM entrants"
	sqlx += " WHERE EntrantStatus IN (" + strconv.Itoa(STATUSCODES["DNS"]) + "," + strconv.Itoa(STATUSCODES["confirmedDNS"])
	showSignedin := r.FormValue("all") != ""
	if showSignedin {
		sqlx += "," + strconv.Itoa(STATUSCODES["signedin"])
	}
	sqlx += ")"
	sqlx += " ORDER BY RiderLast,RiderFirst"

	fmt.Println(sqlx)
	rows, err := DBH.Query(sqlx)
	if err != nil {
		fmt.Println(sqlx)
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)
	fmt.Fprint(w, `<main class="signin">`)

	fmt.Fprint(w, `<h2>RBLR1000 - Signing-in</h2>`)

	fmt.Fprint(w, `<div id="signinlist">`)
	n := 0
	oe := true
	for rows.Next() {
		var e Entrant
		err := rows.Scan(&e.EntrantID, &e.Rider.First, &e.Rider.Last, &e.Rider.IBA, &e.Rider.RBLR, &e.Rider.Email, &e.Rider.Phone, &e.Rider.Address, &e.Pillion.First, &e.Pillion.Last, &e.Pillion.IBA, &e.Pillion.RBLR, &e.Pillion.Email, &e.Pillion.Phone, &e.Pillion.Address, &e.Bike, &e.BikeReg, &e.NokName, &e.NokRelation, &e.NokPhone, &e.OdoStart, &e.StartTime, &e.OdoFinish, &e.FinishTime, &e.EntrantStatus, &e.OdoKms, &e.Route, &e.FundsRaised.EntryDonation, &e.FundsRaised.SquiresCash, &e.FundsRaised.SquiresCheque, &e.FundsRaised.RBLRAccount, &e.FundsRaised.JustGivingAmt)

		checkerr(err)

		//fmt.Printf(`<tr><td>%v</td><td>--%v</td><td>%v</td></tr>`, e.EntrantID, e.Rider.First, e.Rider.Last)
		//fmt.Fprintf(w, `<tr><td>%v</td><td>--%v</td><td>%v</td></tr>`, e.EntrantID, e.Rider.First, e.Rider.Last)

		fmt.Fprint(w, `<div class="signinrow `)
		if oe {
			fmt.Fprint(w, "odd")
		} else {
			fmt.Fprint(w, "even")
		}
		oe = !oe
		fmt.Fprint(w, `">`)

		val, ok := scv[e.EntrantStatus]
		if !ok {
			val = "!" + strconv.Itoa(e.EntrantStatus)
		}

		fmt.Fprintf(w, `<span class="name"><strong>%v</strong>, %v</span> <span class="status">%v</span>`, e.Rider.Last, e.Rider.First, val)

		n++
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div>`)
	fmt.Fprint(w, `<hr><button class="nav" onclick="loadPage('menu');">Main menu</button>`)

	fmt.Fprint(w, `</main></body></html>`)
	fmt.Printf("Showed %v lines\n", n)
}
