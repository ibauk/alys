package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

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
	stt := res[0:11] + st
	//fmt.Println("Starting at " + stt)
	mins := getIntegerFromDB("SELECT StartCohortMins FROM config", 10)
	xtra := getIntegerFromDB("SELECT ExtraCohorts FROM config", 3)
	if stt < res && mins > 0 { // Current time is later than the start time

		for {
			if xtra < 1 {
				break
			}
			if stt >= res {
				break
			}

			// add mins to st
			t, _ := time.ParseInLocation(timefmt, stt, timezone)
			nt := t.Add(time.Minute * time.Duration(mins))
			stt = storeTimeDB(nt)
			//fmt.Println("stt==" + stt)
			st = stt[11:]
			xtra--
		}
	}
	res = res[0:11] + st
	return res, mins, xtra

}

func show_odo(w http.ResponseWriter, r *http.Request, showstart bool) {

	if r.FormValue("debug") != "" {
		fmt.Println("show_odo called")
	}

	sqlx := "SELECT EntrantID,RiderFirst,RiderLast,ifnull(OdoStart,''),ifnull(StartTime,''),ifnull(OdoFinish,''),ifnull(FinishTime,''),EntrantStatus,OdoCounts"
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
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

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

	fmt.Fprintf(w, ` data-gap="%v" data-xtra="%v"`, gap, xtra) // Only needed at start but referenced during timer ticks

	fmt.Fprintf(w, ` >%v</span>`, st[11:16])

	fmt.Fprint(w, ` <span id="ticker">&diams;</span>`)
	if showstart && xtra > 0 {
		fmt.Fprint(w, ` <button onclick="nextTimeSlot();" id="nextSlot"></button>`)
	} else if !showstart {
		const holdlit = `stop clock`
		const unholdlit = `restart clock`
		fmt.Fprintf(w, ` <button data-hold="%v" data-unhold="%v" onclick="clickTimeBtn(this);" id="pauseTime">%v</button>`, holdlit, unholdlit, holdlit)
	}
	fmt.Fprint(w, `<script>`+timerticker+`</script>`)

	fmt.Fprint(w, ` <span id="errlog"></span>`) // Diags only
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<script>setTimeout(reloadPage,30000);refreshTime(); timertick = setInterval(refreshTime,1000);</script>`)

	fmt.Fprint(w, `<div id="odolist">`)
	oe := true
	itemno := 0
	for rows.Next() {
		var EntrantID int
		var RiderFirst, RiderLast, OdoStart, StartTime, OdoFinish, FinishTime string
		var EntrantStatus int
		var OdoCounts string
		rows.Scan(&EntrantID, &RiderFirst, &RiderLast, &OdoStart, &StartTime, &OdoFinish, &FinishTime, &EntrantStatus, &OdoCounts)
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
		fmt.Fprintf(w, `<span><input id="%v" data-e="%v" data-st="%v" name="%v" type="number" class="bignumber" oninput="oi(this);" onchange="oc(this);" min="0" placeholder="%v" value="%v"></span>`, itemno, EntrantID, StartTime, odoname, pch, val)
		fmt.Fprint(w, `</div>`)

	}
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)

	fmt.Fprint(w, `</div><footer><button class="nav" onclick="loadPage('menu');">Main menu</button></footer></body></html>`)
}

func update_odo(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("Here we go")
	if r.FormValue("e") == "" || r.FormValue("f") == "" || r.FormValue("v") == "" {
		fmt.Fprint(w, `{"err":false,"msg":"ok"}`)
		return
	}

	dt := r.FormValue("t")
	if dt == "" {
		dt = storeTimeDB(time.Now())
	}
	sqlx := ""
	switch r.FormValue("f") {
	case "f":
		sqlx = "OdoFinish=" + r.FormValue("v")

		sqlx += ",CorrectedMiles=(" + r.FormValue("v") + " - IfNull(OdoStart,0))"

		ns := STATUSCODES["finishedOK"]
		n, _ := strconv.Atoi(r.FormValue("v"))
		if n < 1 {
			ns = STATUSCODES["DNF"]
			sqlx += ",CertificateAvailable='N'"
		} else if beyond24(r.FormValue("st"), dt) {
			ns = STATUSCODES["finished24+"]
			sqlx += ",CertificateAvailable='N'"
		}

		sqlx += ",FinishTime='" + dt + "'"
		sqlx += ",EntrantStatus=" + strconv.Itoa(ns)
		sqlx += " WHERE EntrantID=" + r.FormValue("e")
		sqlx += " AND FinishTime IS NULL"
		sqlx += " AND EntrantStatus IN (" + strconv.Itoa(STATUSCODES["riding"]) + "," + strconv.Itoa(STATUSCODES["DNF"]) + ")"
	case "s":
		sqlx = "OdoStart=" + r.FormValue("v")
		sqlx += ",StartTime='" + dt + "'"
		sqlx += ",EntrantStatus=" + strconv.Itoa(STATUSCODES["riding"])
		sqlx += " WHERE EntrantID=" + r.FormValue("e")
		sqlx += " AND EntrantStatus IN (" + strconv.Itoa(STATUSCODES["signedin"]) + "," + strconv.Itoa(STATUSCODES["riding"]) + ")"
	}
	fmt.Println(sqlx)
	DBH.Exec("UPDATE entrants SET " + sqlx)

	fmt.Fprint(w, `{"err":false,"msg":"ok"}`)

}
