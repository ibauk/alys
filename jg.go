package main

/*
 * jg.go - handler for JustGiving links and contributions
 *
 * Entrants can use JustGiving pages to help raise donations for the charity and in some cases
 * several entrants will share the same JG page.
 *
 * Code here helps with the URLs. The important part of the url is the "pageShortName", often
 * the name of the entrant or similar unique identifier, but we tend to be given the entire
 * clickable link. The whole url is not suitable for our purposes as it might include variable
 * parameters and/or might follow a different root path to the JG website.
 *
 * Periodically, I circle round the list of pages gathering the latest fundraising info and updating
 * the entrant records so that we're always showing the latest info.
 */

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

const AppID = "48326b85"

const NullCharityReg = "000000"

type jgCharityDetails struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	RegNum string `json:"registrationNumber"`
}

type jgPageDetails struct {
	PageShortName string `json:"pageShortName"`
	Status        string `json:"status"`
	Title         string `json:"title"`
	FundsRaised   string `json:"grandTotalRaisedExcludingGiftAid"`
	NumDonors     string `json:"donationCount"`
	Charity       jgCharityDetails
}

type jgRecord struct {
	PageShortName string
	NumUsers      int
	FundsRaised   int
	PerUser       int
	PageValid     int
	CharityReg    string
	CharityName   string
}

func doJGTest(w http.ResponseWriter, r *http.Request) {

	rebuildJGPages()
}

type pair struct {
	url string
	eid int
}
type justg struct {
	url string
	nu  int
	fr  int
	pu  int
	ok  int
}

func extractJGPSN(w http.ResponseWriter, r *http.Request) {

	jgurl := r.FormValue("jgurl")
	if jgurl == "" {
		fmt.Fprint(w, `{"err":true,"msg":""}`)
		return
	}
	jgpsn := parseJGPageShortName(jgurl)
	fmt.Fprintf(w, `{"err":false,"msg":"%v"}`, jgpsn)
}

// rebuildJGPages rebuilds the table of JustGiving pages starting with examining
// entrant records for a list of PSNs, using the api to fetch data for those PSNs
// and finally calling refreshJGPageTotals to update the entrant records
func rebuildJGPages() {

	fmt.Println("rebuildJGPages()")

	rcsok := strings.Split(getStringFromDB("SELECT JustGCharities FROM config", ""), ",")

	sqlx := "SELECT EntrantID,JustGivingPSN FROM entrants WHERE ifnull(JustGivingPSN,'')<>''"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	psns := make([]pair, 0)
	for rows.Next() {
		var rec pair
		err = rows.Scan(&rec.eid, &rec.url)
		checkerr(err)
		psns = append(psns, rec)
	}
	rows.Close()

	_, err = DBH.Exec("BEGIN TRANSACTION")
	checkerr(err)
	defer DBH.Exec("ROLLBACK")

	sqlx = "DELETE FROM justgs"
	_, err = DBH.Exec(sqlx)
	checkerr(err)

	totfunds := 0
	for _, r := range psns {
		psn := r.url
		pd := getJGPageDetails(psn)
		val := intval(pd.FundsRaised)
		pageValid := 1
		if !slices.Contains(rcsok, pd.Charity.RegNum) {
			pageValid = 0
		} else {
			totfunds += val
		}
		sqlx = fmt.Sprintf("INSERT INTO justgs(PageShortName,NumUsers,CharityReg,CharityName,PageValid,FundsRaised,PerUser)VALUES('%v',1,'%v','%v',%v,%v,%v) ON CONFLICT(PageShortName) DO UPDATE SET NumUsers=NumUsers+1", psn, pd.Charity.RegNum, safesql(pd.Charity.Name), pageValid, val, val)
		//fmt.Println(sqlx)
		_, err = DBH.Exec(sqlx)
		checkerr(err)
	}
	sqlx = "UPDATE justgs SET PerUser=FundsRaised / NumUsers WHERE NumUsers > 1"
	_, err = DBH.Exec(sqlx)
	checkerr(err)
	_, err = DBH.Exec("COMMIT")
	checkerr(err)

	refreshJGPageTotals()
}

// refreshJGPageTotals updates the entrant records using data held in the justgs table
func refreshJGPageTotals() {

	fmt.Println("refreshJGPageTotals()")

	sqlx := "SELECT PageShortName,NumUsers,FundsRaised,PerUser,PageValid FROM justgs"
	urls := make([]justg, 0)
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	for rows.Next() {
		var r justg
		err = rows.Scan(&r.url, &r.nu, &r.fr, &r.pu, &r.ok)
		checkerr(err)
		urls = append(urls, r)
	}
	rows.Close()
	_, err = DBH.Exec("BEGIN TRANSACTION")
	checkerr(err)
	defer DBH.Exec("COMMIT")

	for i, r := range urls {
		pageValid := r.ok
		if pageValid == 0 {
			urls[i].pu = 0
		}
		sqlx = fmt.Sprintf("UPDATE entrants SET JustGivingAmt='%v' WHERE JustGivingPSN='%v'", urls[i].pu, r.url)
		_, err = DBH.Exec(sqlx)
		checkerr(err)
	}

}

// parseJGPageShortName extracts the PageShortName (PSN) from the full url of a
// JustGiving page link
func parseJGPageShortName(url string) string {

	psn := url
	// drop the parameters
	p := strings.Index(url, "?")
	if p > 0 {
		psn = url[:p]
	}
	psnc := strings.Split(psn, "/")
	n := len(psnc) - 1
	psn = psnc[n]
	n--
	if n >= 0 && psnc[n] == "page" {
		psn = psnc[n] + "/" + psn
	}
	return psn
}

// getJGPageDetails uses the api to obtain accurate current data relating to
// the particular PSN
func getJGPageDetails(jgpage string) jgPageDetails {

	var PD jgPageDetails

	jgurl := fmt.Sprintf("https://api.justgiving.com/v1/fundraising/pages/%v", jgpage)
	client := &http.Client{}
	req, err := http.NewRequest("GET", jgurl, nil)
	checkerr(err)
	// ...
	req.Header.Add("x-api-key", AppID)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	checkerr(err)
	defer resp.Body.Close()

	//var bodyString any
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		checkerr(err)
		//fmt.Println(string(bodyBytes)) //
		err = json.Unmarshal(bodyBytes, &PD)
		checkerr(err)
	} else {
		fmt.Printf("%v returned %v\n", jgpage, resp.StatusCode)
		PD.PageShortName = jgpage
		PD.Charity.RegNum = NullCharityReg
		PD.Charity.Name = "page not found"
	}

	return PD
}

func showJustGPages(w http.ResponseWriter, r *http.Request) {

	fmt.Println("showJustGPages()")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	fmt.Fprint(w, refresher)
	fmt.Fprint(w, `<div class="top"><h2 class="link" onclick="loadPage('showjg?refresh=1');">JustGiving pages`)
	fmt.Fprint(w, ` <button onclick="loadPage('showjg?refresh=1');">Refresh pages</button> <span id="jgtotal"></span></h2>`)
	fmt.Fprint(w, `</div>`)
	fmt.Fprint(w, `<main class="justgpages">`)

	if r.FormValue("refresh") != "" {
		fmt.Fprint(w, `<p>refreshing ...</p><p>This will take a few minutes. Please be patient.</p>`)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		rebuildJGPages()
		fmt.Fprint(w, `<script>function x() {loadPage('showjg');}setTimeout(x,0);</script></main></body></html>`)
	}
	sqlx := "SELECT PageShortName,NumUsers,FundsRaised,PerUser,PageValid,CharityReg,CharityName FROM justgs ORDER BY PageShortName"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	oe := true
	totfunds := 0
	for rows.Next() {
		var rec jgRecord
		err = rows.Scan(&rec.PageShortName, &rec.NumUsers, &rec.FundsRaised, &rec.PerUser, &rec.PageValid, &rec.CharityReg, &rec.CharityName)
		checkerr(err)
		fmt.Fprint(w, `<div class="pagerow `)
		if oe {
			fmt.Fprint(w, ` odd`)
		} else {
			fmt.Fprint(w, ` even`)
		}
		oe = !oe
		fmt.Fprint(w, `">`)
		fmt.Fprintf(w, `<span class="PageShortName" title="%v">%v</span>`, rec.PageShortName, rec.PageShortName)
		fmt.Fprintf(w, `<span class="NumUsers"><a href="/search?q=%v"> %v </a></span>`, rec.PageShortName, rec.NumUsers)
		fmt.Fprint(w, `<span class="FundsRaised`)
		if rec.PageValid != 1 {
			fmt.Fprint(w, ` duff`)
		} else {
			totfunds += rec.FundsRaised
		}

		fmt.Fprintf(w, `">£%v`, rec.FundsRaised)
		if rec.NumUsers > 1 {
			fmt.Fprintf(w, ` (%vea)`, rec.PerUser)
		}
		fmt.Fprint(w, `</span>`)
		fmt.Fprint(w, `<span class="RegCharity`)
		if rec.PageValid != 1 {
			fmt.Fprint(w, ` duff`)
		}
		fmt.Fprintf(w, `">%v %v</span>`, rec.CharityReg, rec.CharityName)
		fmt.Fprint(w, `</div>`)
	}

	fmt.Fprint(w, `</main>`)
	fmt.Fprint(w, `<footer><button class="nav" onclick="loadPage('menu');">Main menu</button>  `)
	fmt.Fprint(w, `</footer>`)
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
	fmt.Fprintf(w, `<script>document.getElementById('jgtotal').innerText='£%v';</script>`, totfunds)

	fmt.Fprint(w, `</body></html>`)

}
