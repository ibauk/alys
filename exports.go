package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type JustGivingRec struct {
	FirstName     string
	LastName      string
	Email         string
	Status        string
	JustGivingAmt string
	JustGivingURL string
}

func export_finishers(w http.ResponseWriter, r *http.Request) {

	// I will export records marked as Finisher or Finisher24+
	//
	// The Rides database loader, Rupert, will differentiate between the different classes
	// IBA cert or not. Only 24hr 1000 miles rides will shown on the RoH.

	sqlx := EntrantSQL

	sqlx += " WHERE EntrantStatus=" + strconv.Itoa(STATUSCODES["finishedOK"])
	sqlx += " OR EntrantStatus=" + strconv.Itoa(STATUSCODES["finished24+"])

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=iba1000.json;")

	fmt.Fprintf(w, `{"filetype":"iba1000","asat":"%v","entrants":[`, time.Now().Format(timefmt))

	comma := false
	for rows.Next() {
		var e Entrant

		ScanEntrant(rows, &e)
		b, err := json.Marshal(e)
		checkerr(err)
		if comma {
			fmt.Fprint(w, `,`)
		}
		fmt.Fprintf(w, "%v\n", string(b))
		comma = true
	}
	fmt.Fprint(w, `]}`)
}

func export_JustGiving(w http.ResponseWriter, r *http.Request) {

	var jg = []string{"Rider", "Email", "Ride status", "Amount", "URL"}
	scodes := map[int]string{0: "registered", 1: "withdrawn", 2: "signed-in", 4: "checked-out", 6: "DNF", 8: "Finisher", 10: "Finisher24+"}

	sqlx := EntrantSQL

	sqlx += " WHERE ifnull(JustGivingAmt,'') <> '' OR ifnull(JustGivingURL,'') <> ''"

	//fmt.Println(sqlx)
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=rblr1000jg.csv;")

	jgcsv := csv.NewWriter(w)
	defer jgcsv.Flush()

	err = jgcsv.Write(jg)
	checkerr(err)
	for rows.Next() {
		var e Entrant

		ScanEntrant(rows, &e)

		jg[0] = e.Rider.First + " " + e.Rider.Last
		jg[1] = e.Rider.Email

		es, ok := scodes[e.EntrantStatus]
		if !ok {
			jg[2] = strconv.Itoa(e.EntrantStatus)
		} else {
			jg[2] = es
		}
		jg[3] = e.FundsRaised.JustGivingAmt
		jg[4] = e.FundsRaised.JustGivingURL

		err = jgcsv.Write(jg)
		checkerr(err)
	}

}
