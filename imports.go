package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var ReglistCSV = map[string]int{
	"EntrantID":        0,
	"RiderFirst":       1,
	"RiderLast":        2,
	"RiderIBA":         3,
	"RiderRBL":         4,
	"RiderNovice":      5,
	"PillionFirst":     6,
	"PillionLast":      7,
	"PillionIBA":       8,
	"PillionRBL":       9,
	"PillionNovice":    10,
	"Bike":             11,
	"BikeMake":         12,
	"BikeModel":        13,
	"BikeReg":          14,
	"OdoKms":           15,
	"Email":            16,
	"Phone":            17,
	"Address1":         18,
	"Address2":         19,
	"Town":             20,
	"County":           21,
	"Postcode":         22,
	"Country":          23,
	"NokName":          24,
	"NokPhone":         25,
	"NokRelation":      26,
	"BonusClaimMethod": 27,
	"RouteClass":       28,
	"Tshirt1":          29,
	"Tshirt2":          30,
	"Patches":          31,
	"Camping":          32,
	"Miles2Squires":    33,
	"EnteredDate":      34,
	"PEmail":           35,
	"PPhone":           36,
	"PAddress1":        37,
	"PAddress2":        38,
	"PTown":            39,
	"PCounty":          40,
	"PPostcode":        41,
	"PCountry":         42,
	"Sponsorship":      43,
}

const InsertEntrantSQL = `
	INSERT OR IGNORE INTO entrants (
	EntrantID,Bike,BikeReg,RiderFirst,RiderLast,
	RiderAddress1,RiderAddress2,RiderTown,RiderCounty,
	RiderPostcode,RiderCountry,RiderIBA,RiderPhone,RiderEmail,
	PillionFirst,PillionLast,
	PillionAddress1,PillionAddress2,PillionTown,PillionCounty,
	PillionPostcode,PillionCountry,PillionIBA,PillionPhone,PillionEmail,
	OdoCounts,EntrantStatus,NokName,NokPhone,NokRelation,
	EntryDonation,Route,RiderRBL,PillionRBL,
	Tshirt1,Tshirt2,Patches,FreeCamping,
	CertificateAvailable,CertificateDelivered
	) VALUES (
	?,?,?,?,?,
	?,?,?,?,
	?,?,?,?,?,
	?,?,
	?,?,?,?,
	?,?,?,?,?,
	?,?,?,?,?,
	?,?,?,?,
	?,?,?,?,
	?,?
	 )
`

type EntrantDBRecord struct {
	EntrantID            int
	Bike                 string
	BikeReg              string
	RiderFirst           string
	RiderLast            string
	RiderAddress1        string
	RiderAddress2        string
	RiderTown            string
	RiderCounty          string
	RiderPostcode        string
	RiderCountry         string
	RiderIBA             string
	RiderPhone           string
	RiderEmail           string
	PillionFirst         string
	PillionLast          string
	PillionAddress1      string
	PillionAddress2      string
	PillionTown          string
	PillionCounty        string
	PillionPostcode      string
	PillionCountry       string
	PillionIBA           string
	PillionPhone         string
	PillionEmail         string
	OdoCounts            string
	OdoStart             string
	OdoFinish            string
	CorrectedMiles       string
	FinishTime           string
	StartTime            string
	EntrantStatus        int
	NokName              string
	NokPhone             string
	NokRelation          string
	EntryDonation        string
	SquiresCheque        string
	SquiresCash          string
	RBLRAccount          string
	JustGivingAmt        string
	JustGivingURL        string
	Route                string
	RiderRBL             string
	PillionRBL           string
	Tshirt1              string
	Tshirt2              string
	Patches              int
	FreeCamping          string
	CertificateDelivered string
	CertificateAvailable string
}

func intval(x string) int {
	res := 0
	for i := 0; i < len(x); i++ {
		n := x[i]
		if n >= '0' && n <= '9' {
			res = res * 10
			res = res + int(n) - int('0')
		} else if n != '£' {
			break
		}
	}
	return res
}

func isValidReglistData(hdr []string) bool {

	re := regexp.MustCompile(`^Entrantid,RiderFirst,RiderLast.*RouteClass.*Miles2Squires`)
	return re.MatchString(strings.Join(hdr, ","))
}

func load_entrants(w http.ResponseWriter, r *http.Request) {

	zap := r.FormValue("updatemode") == "overwrite"
	certavail := r.FormValue("certavail")
	if certavail == "" {
		certavail = "Y"
	}

	n := LoadEntrantsFromCSV(r.FormValue("csvdata"), certavail, zap, true)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<main class="upload">`)
	fmt.Fprint(w, `<h2>Entrant upload complete</h2>`)

	if n < 1 {
		fmt.Fprint(w, `<p>No entrants were loaded</p>`)
	} else {
		if zap {
			fmt.Fprint(w, `<p>This upload replaced the entire set of entrants</p>`)
		} else {
			fmt.Fprint(w, `<p>New entrants were added to the existing dataset</p>`)
		}

		if n == 1 {
			fmt.Fprintf(w, `<p>A single, wafer-thin, entrant was loaded. Certificate available=%v</p>`, certavail)
		} else {
			fmt.Fprintf(w, `<p>%v entrants loaded. Certificates available=%v</p>`, n, certavail)
		}
	}

	rex := getIntegerFromDB("SELECT count(*) FROM entrants", -1)
	fmt.Fprintf(w, `<p>The database now contains %v entrant records</p>`, rex)

	fmt.Fprint(w, `</main>`)
	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
	fmt.Fprint(w, `<footer>`)
	fmt.Fprint(w, `<button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, `</footer>`)

	fmt.Fprint(w, `</body><html>`)

}
func LoadEntrantsFromCSV(csvFile string, certAvail string, zapExisting bool, fileIsData bool) int64 {

	var rex [][]string
	if fileIsData {
		rex = readCsvData(csvFile)
	} else {
		rex = readCsvFile(csvFile)
	}

	stmt, err := DBH.Prepare(InsertEntrantSQL)
	checkerr(err)
	defer stmt.Close()
	n := 0
	nl := int64(0)
	r := ReglistCSV // convenient shorthand
	for _, ln := range rex {
		n++
		if n == 1 {
			if !isValidReglistData(ln) {
				fmt.Println("Import file not valid")
				return 0
			}
			if zapExisting {
				_, err := DBH.Exec("DELETE FROM entrants")
				checkerr(err)
			}

			_, err := DBH.Exec("BEGIN TRANSACTION")
			checkerr(err)

			continue
		}
		//fmt.Printf("%v\n", ln[r["RiderLast"]])
		patches := intval(ln[r["Patches"]])
		if err != nil {
			fmt.Printf("%v gives %v with error %v\n", ln[r["Patches"]], patches, err)
		}
		fmt.Printf("%v={%v,'%v'}    ", ln[r["EntrantID"]], ln[r["OdoKms"]], ln[r["RiderRBL"]])
		if true {
			res, err := stmt.Exec(ln[r["EntrantID"]], ln[r["Bike"]], ln[r["BikeReg"]], ln[r["RiderFirst"]], ln[r["RiderLast"]],
				ln[r["Address1"]], ln[r["Address2"]], ln[r["Town"]], ln[r["County"]],
				ln[r["Postcode"]], ln[r["Country"]], ln[r["RiderIBA"]], ln[r["Phone"]], ln[r["Email"]],
				ln[r["PillionFirst"]], ln[r["PillionLast"]],
				ln[r["PAddress1"]], ln[r["PAddress2"]], ln[r["PTown"]], ln[r["PCounty"]],
				ln[r["PPostcode"]], ln[r["PCountry"]], ln[r["PillionIBA"]], ln[r["PPhone"]], ln[r["PEmail"]],
				ln[r["OdoKms"]], STATUSCODES["DNS"], ln[r["NokName"]], ln[r["NokPhone"]], ln[r["NokRelation"]],
				ln[r["Sponsorship"]], RouteClass(ln[r["RouteClass"]]), ln[r["RiderRBL"]], ln[r["PillionRBL"]],
				ln[r["Tshirt1"]], ln[r["Tshirt2"]], patches, ln[r["FreeCamping"]],
				certAvail, "N",
			)
			checkerr(err)
			ra, err := res.RowsAffected()
			checkerr(err)
			nl += ra
		}
	}
	DBH.Exec("COMMIT")
	fmt.Printf("%v records loaded\n", nl)
	return nl
}

func readCsvData(filedata string) [][]string {

	csvReader := csv.NewReader(strings.NewReader(filedata))
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV ", err)
	}

	return records
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func show_loadCSV(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, refresher)

	fmt.Fprint(w, `<main class="upload">`)
	fmt.Fprint(w, `<h2>Import Entrants</h2>`)
	fmt.Fprint(w, `<p>Import entrant details from a CSV file prepared by <em>Reglist</em>.</p>`)
	fmt.Fprint(w, `<form action="/upload" method="post" enctype="multipart/form-data" onsubmit="importEntrantsCSV(this)">`)

	fmt.Fprint(w, `<fieldset>`)
	fmt.Fprint(w, `<label for="updatemode">Update mode</label> `)
	fmt.Fprint(w, `<select id="updatemode" name="updatemode">`)
	fmt.Fprint(w, `<option selected value="append">Add new entries only</option>`)
	fmt.Fprint(w, `<option value="overwrite">Replace all entries</option>`)
	fmt.Fprint(w, `</select>`)
	fmt.Fprint(w, `</fieldset>`)

	fmt.Fprint(w, `<fieldset>`)
	fmt.Fprint(w, `<label for="certavail">Certificates available</label> `)
	fmt.Fprint(w, `<select id="certavail" name="certavail">`)
	fmt.Fprint(w, `<option selected value="Y">Yes, already printed</option>`)
	fmt.Fprint(w, `<option value="N">No, printing needed</option>`)
	fmt.Fprint(w, `</select>`)
	fmt.Fprint(w, `</fieldset>`)

	fmt.Fprint(w, `<fieldset>`)
	fmt.Fprint(w, `<label for="csvfile">CSV (reglist) file to upload</label> `)
	fmt.Fprint(w, `<input id="csvfile" name="csvfile" type="file" accept=".csv" onchange="enableImportLoad(this)">`)
	fmt.Fprint(w, `</fieldset>`)

	fmt.Fprint(w, `<input type="hidden" id="csvdata" name="csvdata" value="nodata">`)

	fmt.Fprint(w, `<fieldset id="importloader" class="hide"><button>Update database</button></fieldlist>`)
	fmt.Fprint(w, `</form>`)
	fmt.Fprint(w, `</main>`)

	fmt.Fprint(w, `<script>document.onkeydown=function(e){if(e.keyCode==27) {e.preventDefault();loadPage('menu');}}</script>`)
	fmt.Fprint(w, `<footer>`)
	fmt.Fprint(w, `<button class="nav" onclick="loadPage('menu');">Main menu</button>`)
	fmt.Fprint(w, `</footer>`)

	fmt.Fprint(w, `</body></html>`)
}
