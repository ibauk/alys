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

func put_odo_update(sqlx string) {

	mysqlx := "UPDATE entrants SET " + sqlx

	fmt.Println(mysqlx)
	_, err := DBH.Exec(mysqlx)
	checkerr(err)
}
func show_odo(w http.ResponseWriter, r *http.Request, showstart bool, fullaccess bool) {

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

	//	const MinsB4Reload = 10
	//	const SecsB4Reload = 0
	//	const MsecsB4Reload = ((MinsB4Reload * 60) + SecsB4Reload) * 1000
	fmt.Fprint(w, `<script>refreshTime(); timertick = setInterval(refreshTime,1000);</script>`)

	fmt.Fprint(w, `<div id="odolist">`)
	oe := true
	itemno := 0
	minOdoDiff := getIntegerFromDB("SELECT MinOdoDiff FROM config", 0)
	maxOdoDiff := getIntegerFromDB("SELECT MaxOdoDiff FROM config", 0)
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

		fmt.Fprintf(w, `<span class="name"><label for="%v"><strong>%v</strong>, %v</label></span> `, itemno, RiderLast, RiderFirst)
		pch := "finish odo"
		val := OdoFinish
		if showstart {
			pch = "start odo"
			val = OdoStart
		}
		fmt.Fprintf(w, `<span><input id="%v" data-e="%v" data-st="%v" data-so="%v" data-oc="%v" name="%v" type="number" class="bignumber" oninput="oi(this);" onchange="oc(this);" onblur="oc(this);" min="0" placeholder="%v" value="%v" data-minod="%v" data-maxod="%v" ondblclick="explainOdo(this);" autocomplete="off"></span>`, itemno, EntrantID, StartTime, OdoStart, OdoCounts, odoname, pch, val, minOdoDiff, maxOdoDiff)
		fmt.Fprint(w, `</div>`)

	}

	if fullaccess {
		fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
		fmt.Fprint(w, `</div><footer><button class="nav" onclick="loadPage('menu');">Main menu</button></footer></body></html>`)
	} else {
		fmt.Fprint(w, `</div><footer><button class="nav" onclick="loadPage('stats');">Live stats</button></footer></body></html>`)
	}
}

// update_odo updates a start or finish odo reading and also updates the entrant
// status value and start/finish times.
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

	// First we update the odo reading without altering state
	if r.FormValue("f") == "f" {
		sqlx = "OdoFinish=" + r.FormValue("v")
		// This needs to cater for odos reflecting M/K flag
		sqlx += ",CorrectedMiles=(" + r.FormValue("v") + " - IfNull(OdoStart,0))"
	} else {
		sqlx = "OdoStart=" + r.FormValue("v")
	}
	sqlx += " WHERE EntrantID=" + r.FormValue("e")
	put_odo_update(sqlx)

	sqlx = ""
	switch r.FormValue("f") {
	case "f":
		sqlx += "FinishTime='" + dt + "'"

		ns := STATUSCODES["finishedOK"]
		n, _ := strconv.Atoi(r.FormValue("v"))
		if n < 1 {
			ns = STATUSCODES["DNF"]
			sqlx += ",CertificateAvailable='N'"
		} else if beyond24(r.FormValue("st"), dt) {
			ns = STATUSCODES["finished24+"]
			sqlx += ",CertificateAvailable='N'"
		}
		sqlx += ",EntrantStatus=" + strconv.Itoa(ns)

		wherex := " WHERE EntrantID=" + r.FormValue("e")
		wherex += " AND ifnull(FinishTime,'')=''"

		statusx := " AND EntrantStatus IN (" + strconv.Itoa(STATUSCODES["riding"]) + "," + strconv.Itoa(STATUSCODES["DNF"]) + ")"

		put_odo_update(sqlx + wherex + statusx)

	case "s":
		sqlx += "StartTime='" + dt + "'"
		sqlx += ",EntrantStatus=" + strconv.Itoa(STATUSCODES["riding"])
		sqlx += " WHERE EntrantID=" + r.FormValue("e")
		sqlx += " AND EntrantStatus IN (" + strconv.Itoa(STATUSCODES["signedin"]) + "," + strconv.Itoa(STATUSCODES["riding"]) + ")"
		put_odo_update(sqlx)
	}

	fmt.Fprint(w, `{"err":false,"msg":"ok"}`)

}
