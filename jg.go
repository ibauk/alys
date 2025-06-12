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
	"strings"
)

const AppID = "48326b85"
const testPage = "Pawel-Janik" //"jason-bassett-1" //"rblr10002025" //"jason-bassett-1"

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

var PD jgPageDetails

type donation struct {
	Amount string `json:"amount"`
}
type pagination struct {
	PageSizeRequested int
	PageSizeReturned  int `json:"pageSizeReturned"`
	TotalPages        int
	TotalResults      int
}

/*
*

	"id": "page/rblr10002025",
	"pageShortName": "page/rblr10002025",
	"pagination": {
	  "nextPageCursor": null,
	  "pageSizeRequested": 50,
	  "pageSizeReturned": 40,
	  "totalPages": 1,
	  "totalResults": 40
	}

*
*/
type donations struct {
	Donations     []donation `json:"donations"`
	Id            string
	PageShortName string
	Pagination    pagination `json:"pagination"`
}

var DD donations

/**
	{
      "amount": "30",
      "charityId": 250914,
      "currencyCode": "GBP",
      "donationDate": "/Date(1749476949000)/",
      "donationRef": null,
      "donorDisplayName": "Anonymous",
      "donorLocalAmount": "30",
      "donorLocalCurrencyCode": "GBP",
      "estimatedTaxReclaim": 7.5,
      "id": 1147789313,
      "image": "https://www.justgiving.com/content/images/graphics/icons/avatars/facebook-avatar.gif",
      "message": "Fantastic effort guys! All completed, even if a bit soggy and tired! We’ll done all!",
      "source": "SponsorshipDonations",
      "thirdPartyReference": null
    }
**/

var urls = []string{
	"https://www.justgiving.com/page/rblr1000?utm_medium=FR&utm_source=CL&utm_campaign=015",
	"https://www.justgiving.com/page/dave-broome-2?utm_medium=FR&utm_source=CL&utm_campaign=015",
	"https://www.justgiving.com/page/rblr1000-smiths-1735650564643?utm_medium=FR&utm_source=CL&utm_campaign=015",
	"https://www.justgiving.com/fundraising/Pawel-Janik?utm_medium=FR&utm_source=CL&utm_campaign=015",
	"https://www.justgiving.com/page/rblr10002025",
}

func doJGTest(w http.ResponseWriter, r *http.Request) {

	getFundsRaised(testPage)

	fmt.Fprintln(w, "<p>Parsing urls</p>")
	for u := range urls {
		psn := parsePageShortName(urls[u])
		fmt.Fprintf(w, "%v == %v\n", psn, urls[u])
		fmt.Fprintf(w, "<p>Funds raised %v</p>", getFundsRaised(psn))
	}

}
func doJGTestOffline() {

	getFundsRaised(testPage)

	fmt.Println("Parsing urls")
	for u := range urls {
		psn := parsePageShortName(urls[u])
		fmt.Printf("%v == %v\n", psn, urls[u])
		getFundsRaised(psn)
	}

}
func parsePageShortName(url string) string {

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
func getFundsRaised(jgpage string) int {

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
	}

	fmt.Printf("%v == %v \n", jgurl, resp.Status)

	totval := intval(PD.FundsRaised)
	fmt.Printf("Total funds: %v\n", totval)

	//fmt.Printf("Page details\n%v\n", PD)
	// ...

	return totval

}
