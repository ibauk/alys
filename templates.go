package main

type Person = struct {
	First   string
	Last    string
	IBA     string
	RBLR    string
	Email   string
	Phone   string
	Address string
}

type Money = struct {
	EntryDonation string
	SquiresCheque string
	SquiresCash   string
	RBLRAccount   string
	JustGivingAmt string
}

type Entrant = struct {
	EntrantID     int
	EntrantStatus int
	Rider         Person
	Pillion       Person
	NokName       string
	NokRelation   string
	NokPhone      string
	Bike          string
	BikeReg       string
	Route         string
	OdoStart      string
	OdoFinish     string
	OdoKms        int
	StartTime     string
	FinishTime    string
	FundsRaised   Money
}

const SigninScreenSingle = `
<div class="SigninScreenSingle">
<label for="RiderFirst">First name</label> <input id="RiderFirst" name="RiderFirst" class="RiderFirst" value="{{.RiderFirst}}">
</div>
`
