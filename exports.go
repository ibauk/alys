package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func export_finishers(w http.ResponseWriter, r *http.Request) {

	// I will export records marked as Finisher, not Finisher24+, where CorrectedMiles > 1000

	sqlx := EntrantSQL

	sqlx += " WHERE EntrantStatus=" + strconv.Itoa(STATUSCODES["finishedOK"])
	sqlx += " AND CorrectedMiles >=1000"

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
