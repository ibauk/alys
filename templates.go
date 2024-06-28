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
<fieldset class="tabContent" id="tab_rider"><legend>Rider</legend>
<div class="field"><div class="field"><label for="RiderLast">Last name</label> <input id="RiderLast" name="RiderLast" class="RiderLast" value="{{.Rider.Last}}"></div>
<div class="field"><label for="RiderFirst">First name</label> <input id="RiderFirst" name="RiderFirst" class="RiderFirst" value="{{.Rider.First}}"></div>
<div class="field"><label for="RiderIBA">IBA #</label> <input id="RiderIBA" name="RiderIBA" class="RiderIBA" value="{{.Rider.IBA}}"></div>
<div class="field"><label for="RiderRBLR">RBL Member</label> <input id="RiderRBLR" name="RiderRBLR" class="RiderRBLR" value="{{.Rider.RBLR}}"></div>
<div class="field"><label for="RiderEmail">Email</label> <input id="RiderEmail" name="RiderEmail" class="RiderEmail" value="{{.Rider.Email}}"></div>
<div class="field"><label for="RiderPhone">Mobile</label> <input id="RiderPhone" name="RiderPhone" class="RiderPhone" value="{{.Rider.Phone}}"></div>
<div class="field"><label for="RiderAddress">Address</label> <input id="RiderAddress" name="RiderAddress" class="RiderAddress" value="{{.Rider.Address}}"></div>
<div class="field">
    <label for="Route">Route</label> 
	<select id="Route" name="Route">
	    <option value="A-NCW"{{if eq .Route "A-NCW"}} selected{{end}}>North clockwise</option>
	    <option value="B-NAC"{{if eq .Route "B-NAC"}} selected{{end}}>North anticlockwise</option>
	    <option value="C-SCW"{{if eq .Route "C-SCW"}} selected{{end}}>South clockwise</option>
	    <option value="D-SAC"{{if eq .Route "D-SAC"}} selected{{end}}>South anticlockwise</option>
	    <option value="E-500CW"{{if eq .Route "E-500CW"}} selected{{end}}>500 clockwise</option>
	    <option value="F-500AC"{{if eq .Route "F-500AC"}} selected{{end}}>500 anticlockwise</option>
	</select>
</div>
<div class="field">
    <label for="EntrantStatus" name="EntrantStatus">Status</label>
	<select id="EntrantStatus" name="EntrantStatus">
	    <option value="0"{{if eq .EntrantStatus 0}} selected{{end}}>not signed in</option>
	    <option value="1"{{if eq .EntrantStatus 1}} selected{{end}}>withdrawn</option>
	    <option value="2"{{if eq .EntrantStatus 2}} selected{{end}}>signed in</option>
	    <option value="4"{{if eq .EntrantStatus 4}} selected{{end}}>checked out</option>
	    <option value="6"{{if eq .EntrantStatus 6}} selected{{end}}>DNF</option>
	    <option value="8"{{if eq .EntrantStatus 8}} selected{{end}}>Finisher</option>
	    <option value="10"{{if eq .EntrantStatus 10}} selected{{end}}>Late finisher</option>
	</select>
</div>
</fieldset>
<fieldset class="tabContent" id="tab_bike"><legend>Bike</legend>
<div class="field"><label for="Bike">Bike</label> <input id="Bike" name="Bike" class="Bike" value="{{.Bike}}"></div>
<div class="field"><label for="BikeReg">Registration</label> <input id="BikeReg" name="BikeReg" class="BikeReg" value="{{.BikeReg}}"></div>
<div class="field"><label for="OdoKms">Odo counts</label> <input id="OdoKms" name="OdoKms" class="OdoKms" value="{{.OdoKms}}"></div>
<div class="field">
    <span class="label">Odo counts</span>
	<input type="radio" id="OdoKms0" name="OdoKms" value="0"{{if ne .OdoKms 1}} checked{{end}}> <label for="OdoKms0">miles</label> 
	<input type="radio" id="OdoKms1" name="OdoKms" value="1"{{if eq .OdoKms 1}} checked{{end}}> <label for="OdoKms1">kms</label>
</div>
<div class="field"><label for="OdoStart">Odo @ start</label> <input id="OdoStart" name="OdoStart" class="OdoStart" value="{{.OdoStart}}"></div>
<div class="field"><label for="OdoFinish">Odo @ finish</label> <input id="OdoFinish" name="OdoFinish" class="OdoFinish" value="{{.OdoFinish}}"></div>
</fieldset>
<fieldset class="tabContent" id="tab_nok"><legend>Emergency contact</legend>
<div class="field"><label for="NokName">Contact name</label> <input id="NokName" name="NokName" class="NokName" value="{{.NokName}}"></div>
<div class="field"><label for="NokRelation">Relationship</label> <input id="NokRelation" name="NokRelation" class="NokRelation" value="{{.NokRelation}}"></div>
<div class="field"><label for="NokPhone">Contact phone</label> <input id="NokPhone" name="NokPhone" class="NokPhone" value="{{.NokPhone}}"></div>
</fieldset>
<fieldset class="tabContent" id="tab_pillion"><legend>Pillion</legend>
<div class="field"><label for="PillionLast">Last name</label> <input id="PillionLast" name="PillionLast" class="PillionLast" value="{{.Pillion.Last}}"></div>
<div class="field"><label for="PillionFirst">First name</label> <input id="PillionFirst" name="PillionFirst" class="PillionFirst" value="{{.Pillion.First}}"></div>
<div class="field"><label for="PillionIBA">IBA #</label> <input id="PillionIBA" name="PillionIBA" class="PillionIBA" value="{{.Pillion.IBA}}"></div>
<div class="field"><label for="PillionRBLR">RBL Member</label> <input id="PillionRBLR" name="PillionRBLR" class="PillionRBLR" value="{{.Pillion.RBLR}}"></div>
<div class="field"><label for="PillionEmail">Email</label> <input id="PillionEmail" name="PillionEmail" class="PillionEmail" value="{{.Pillion.Email}}"></div>
<div class="field"><label for="PillionPhone">Mobile</label> <input id="PillionPhone" name="PillionPhone" class="PillionPhone" value="{{.Pillion.Phone}}"></div>
<div class="field"><label for="PillionAddress">Address</label> <input id="PillionAddress" name="PillionAddress" class="PillionAddress" value="{{.Pillion.Address}}"></div>
</fieldset>
<fieldset class="tabContent" id="tab_money"><legend>Money</legend>
<div class="field"><label for="PillionLast">Last name</label> <input id="PillionLast" name="PillionLast" class="PillionLast" value="{{.Pillion.Last}}"></div>
<div class="field"><label for="PillionFirst">First name</label> <input id="PillionFirst" name="PillionFirst" class="PillionFirst" value="{{.Pillion.First}}"></div>
<div class="field"><label for="PillionIBA">IBA #</label> <input id="PillionIBA" name="PillionIBA" class="PillionIBA" value="{{.Pillion.IBA}}"></div>
<div class="field"><label for="PillionRBLR">RBL Member</label> <input id="PillionRBLR" name="PillionRBLR" class="PillionRBLR" value="{{.Pillion.RBLR}}"></div>
<div class="field"><label for="PillionEmail">Email</label> <input id="PillionEmail" name="PillionEmail" class="PillionEmail" value="{{.Pillion.Email}}"></div>
<div class="field"><label for="PillionPhone">Mobile</label> <input id="PillionPhone" name="PillionPhone" class="PillionPhone" value="{{.Pillion.Phone}}"></div>
<div class="field"><label for="PillionAddress">Address</label> <input id="PillionAddress" name="PillionAddress" class="PillionAddress" value="{{.Pillion.Address}}"></div>
</fieldset>



</div>
`
