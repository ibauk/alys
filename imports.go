package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
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
	INSERT INTO entrants (
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
		} else {
			break
		}
	}
	return res
}
func LoadEntrantsFromCSV(csvFile string) {

	rex := readCsvFile(csvFile)
	_, err := DBH.Exec("DELETE FROM entrants")
	checkerr(err)
	stmt, err := DBH.Prepare(InsertEntrantSQL)
	checkerr(err)
	defer stmt.Close()
	n := 0
	r := ReglistCSV // convenient shorthand
	for _, ln := range rex {
		n++
		if n == 1 {
			continue
		}
		fmt.Printf("%v\n", ln[r["RiderLast"]])
		patches := intval(ln[r["Patches"]])
		if err != nil {
			fmt.Printf("%v gives %v with error %v\n", ln[r["Patches"]], patches, err)
		}
		_, err = stmt.Exec(ln[r["EntrantID"]], ln[r["Bike"]], ln[r["BikeReg"]], ln[r["RiderFirst"]], ln[r["RiderLast"]],
			ln[r["Address1"]], ln[r["Address2"]], ln[r["Town"]], ln[r["County"]],
			ln[r["Postcode"]], ln[r["Country"]], ln[r["RiderIBA"]], ln[r["Phone"]], ln[r["Email"]],
			ln[r["PillionFirst"]], ln[r["PillionLast"]],
			ln[r["PAddress1"]], ln[r["PAddress2"]], ln[r["PTown"]], ln[r["PCounty"]],
			ln[r["PPostcode"]], ln[r["PCountry"]], ln[r["PillionIBA"]], ln[r["PPhone"]], ln[r["PEmail"]],
			ln[r["OdoCounts"]], STATUSCODES["DNS"], ln[r["NokName"]], ln[r["NokPhone"]], ln[r["NokRelation"]],
			ln[r["Sponsorship"]], RouteClass(ln[r["RouteClass"]]), ln[r["RiderRBL"]], ln[r["PillionRBL"]],
			ln[r["Tshirt1"]], ln[r["Tshirt2"]], patches, ln[r["FreeCamping"]],
			"Y", "N",
		)
		checkerr(err)
	}
	fmt.Printf("%v records loaded\n", n)
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

func RouteClass(rc string) string {

	RC := map[string]string{
		"A": "A-NCW",
		"B": "B-NAC",
		"C": "C-SCW",
		"D": "D-SAC",
		"E": "E-500CW",
		"F": "F-500AC",
	}
	rca := rc[0:1]
	val, ok := RC[rca]
	if !ok {
		return RC["A"]
	}
	return val
}
