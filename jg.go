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
func rebuildJGPages() {

	rcsok := strings.Split(getStringFromDB("SELECT JustGCharities FROM config", ""), ",")

	sqlx := "SELECT EntrantID,JustGivingURL FROM entrants WHERE ifnull(JustGivingURL,'')<>''"
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

	for _, r := range psns {
		psn := parseJGPageShortName(r.url)
		pd := getJGPageDetails(psn)
		sqlx = fmt.Sprintf("UPDATE entrants SET JustGivingPSN='%v' WHERE EntrantID=%v", psn, r.eid)
		_, err = DBH.Exec(sqlx)
		checkerr(err)
		pageValid := 1
		if !slices.Contains(rcsok, pd.Charity.RegNum) {
			pageValid = 0
		}
		sqlx = fmt.Sprintf("INSERT INTO justgs(PageShortName,NumUsers,CharityReg,CharityName,PageValid)VALUES('%v',1,'%v','%v',%v) ON CONFLICT(PageShortName) DO UPDATE SET NumUsers=NumUsers+1", psn, pd.Charity.RegNum, safesql(pd.Charity.Name), pageValid)
		//fmt.Println(sqlx)
		_, err = DBH.Exec(sqlx)
		checkerr(err)
	}
	_, err = DBH.Exec("COMMIT")
	checkerr(err)

	refreshJGPageTotals()
}

func refreshJGPageTotals() {

	sqlx := "SELECT PageShortName,NumUsers FROM justgs"
	urls := make([]justg, 0)
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	for rows.Next() {
		var r justg
		err = rows.Scan(&r.url, &r.nu)
		checkerr(err)
		urls = append(urls, r)
	}
	rows.Close()
	_, err = DBH.Exec("BEGIN TRANSACTION")
	checkerr(err)
	defer DBH.Exec("COMMIT")

	for i, r := range urls {
		n := getJGFundsRaised(r.url)
		urls[i].fr = n
		urls[i].pu = n
		if r.nu > 1 {
			urls[i].pu = n / r.nu
		}
		if getIntegerFromDB(fmt.Sprintf("SELECT PageValid FROM justgs WHERE PageShortName='%v'", r.url), 0) == 0 {
			urls[i].pu = 0
		}
		sqlx = fmt.Sprintf("UPDATE entrants SET JustGivingAmt='%v' WHERE JustGivingPSN='%v'", urls[i].pu, r.url)
		_, err = DBH.Exec(sqlx)
		checkerr(err)
		sqlx = fmt.Sprintf("UPDATE justgs SET FundsRaised=%v,PerUser=%v WHERE PageShortName='%v'", urls[i].fr, urls[i].pu, r.url)
		_, err = DBH.Exec(sqlx)
		checkerr(err)
	}

}
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
func getJGFundsRaised(jgpage string) int {

	PD := getJGPageDetails(jgpage)
	totval := intval(PD.FundsRaised)

	return totval

}

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
