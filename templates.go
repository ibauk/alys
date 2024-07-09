package main

import (
	"database/sql"
	_ "embed"
)

//go:embed tabs.js
var my_tabs_js string

type ConfigRecord struct {
	StartTime       string
	StartCohortMins int
	ExtraCohorts    int
	RallyStatus     string
}

const ConfigSQL = `SELECT ifnull(StartTime,'05:00'),ifnull(StartCohortMins,10),ifnull(ExtraCohorts,3),ifnull(RallyStatus,'S') FROM config`

const ConfigScreen = `
<div class="ConfigScreen">
	<div class="field">
		<label for="StartTime">Earliest start time</label> 
		<input type="time" class="StartTime" id="StartTime" name="StartTime" value="{{.StartTime}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<label for="StartCohortMins">Minutes between cohorts</label> 
		<input type="number" min="1" max="40" class="StartCohortMins" id="StartCohortMins" name="StartCohortMins" value="{{.StartCohortMins}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<label for="ExtraCohorts">Number of extra cohorts</label> 
		<input type="number" min="0" max="10" class="ExtraCohorts" id="ExtraCohorts" name="ExtraCohorts" value="{{.ExtraCohorts}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<span class="label">State of play: </span>
		<input type="radio" id="RallyStatusS" class="RallyStatus" name="RallyStatus" value="S" {{if ne .RallyStatus "F"}} checked{{end}} data-chg="1" data-static="1" onchange="ocdcfg(this);">
		<label for="RallyStatusS">Signin and start</label>
		<input type="radio" id="RallyStatusF" class="RallyStatus" name="RallyStatus" value="F" {{if eq .RallyStatus "F"}} checked{{end}} data-chg="1" data-static="1" onchange="ocdcfg(this);">
		<label for="RallyStatusF">Check back in and finish</label>
	</div>
</div>
`

type Person = struct {
	First    string
	Last     string
	IBA      string
	RBLR     string
	Email    string
	Phone    string
	Address  string
	Address1 string
	Address2 string
	Town     string
	County   string
	Postcode string
	Country  string
}

type Money = struct {
	EntryDonation string
	SquiresCheque string
	SquiresCash   string
	RBLRAccount   string
	JustGivingAmt string
	JustGivingURL string
}

type Entrant = struct {
	EntrantID            int
	EntrantStatus        int
	Rider                Person
	Pillion              Person
	NokName              string
	NokRelation          string
	NokPhone             string
	Bike                 string
	BikeReg              string
	Route                string
	OdoStart             string
	OdoFinish            string
	OdoCounts            string
	StartTime            string
	FinishTime           string
	FundsRaised          Money
	FreeCamping          string
	CertificateDelivered string
	Tshirt1              string
	Tshirt2              string
	Patches              int
	EditMode             string
}

const EntrantSQL = `SELECT EntrantID,ifnull(RiderFirst,''),ifnull(RiderLast,''),ifnull(RiderIBA,''),ifnull(RiderRBLR,''),ifnull(RiderEmail,''),ifnull(RiderPhone,'')
    ,ifnull(RiderAddress1,''),ifnull(RiderAddress2,''),ifnull(RiderTown,''),ifnull(RiderCounty,''),ifnull(RiderPostcode,''),ifnull(RiderCountry,'')
	,ifnull(PillionFirst,''),ifnull(PillionLast,''),ifnull(PillionIBA,''),ifnull(PillionRBLR,''),ifnull(PillionEmail,''),ifnull(PillionPhone,'')
    ,ifnull(PillionAddress1,''),ifnull(PillionAddress2,''),ifnull(PillionTown,''),ifnull(PillionCounty,''),ifnull(PillionPostcode,''),ifnull(PillionCountry,'')
	,ifnull(Bike,'motorbike'),ifnull(BikeReg,'')
	,ifnull(NokName,''),ifnull(NokRelation,''),ifnull(NokPhone,'')
	,ifnull(OdoStart,''),ifnull(StartTime,''),ifnull(OdoFinish,''),ifnull(FinishTime,''),EntrantStatus,ifnull(OdoCounts,'M'),ifnull(Route,'')
	,ifnull(EntryDonation,''),ifnull(SquiresCash,''),ifnull(SquiresCheque,''),ifnull(RBLRAccount,''),ifnull(JustGivingAmt,'')
	,ifnull(Tshirt1,''),ifnull(Tshirt2,''),ifnull(Patches,0),ifnull(FreeCamping,''),ifnull(CertificateDelivered,'')
	 FROM entrants
`

var SigninScreenSingle = `
<div class="SigninScreenSingle">
<input type="hidden" id="EntrantID" name="EntrantID" value="{{.EntrantID}}">
<input type="hidden" id="EditMode" name="EditMode" value="{{.EditMode}}">
<fieldset class="tabContent" id="tab_rider"><legend>Rider</legend>
<div class="field"><div class="field"><label for="RiderLast">Last name</label> <input autofocus id="RiderLast" name="RiderLast" class="RiderLast" value="{{.Rider.Last}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="RiderFirst">First name</label> <input id="RiderFirst" name="RiderFirst" class="RiderFirst" value="{{.Rider.First}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="RiderIBA">IBA member</label> <input type="checkbox" id="RiderIBA" name="RiderIBA" class="RiderIBA" value="RiderIBA"{{if ne .Rider.IBA ""}} checked{{end}} onchange="oic(this);"></div>
<div class="field"><label for="RiderRBLR">RBL Member</label> <input type="checkbox" id="RiderRBLR" name="RiderRBLR" class="RiderRBLR" value="RiderRBLR"{{if ne .Rider.RBLR ""}} checked{{end}} onchange="oic(this);"></div>
<div class="field"><label for="RiderEmail">Email</label> <input id="RiderEmail" name="RiderEmail" class="RiderEmail" value="{{.Rider.Email}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="RiderPhone">Mobile</label> <input id="RiderPhone" name="RiderPhone" class="RiderPhone" value="{{.Rider.Phone}}" oninput="oid(this);" onchange="ocd(this);"></div>
<br>
<div class="field"> <fieldset><legend class="small">Address</legend>
    <input id="RiderAddress1" name="RiderAddress1" class="RiderAddress1"  value="{{.Rider.Address1}}" oninput="oid(this);" onchange="ocd(this);"><br>
    <input id="RiderAddress2" name="RiderAddress2" class="RiderAddress2"  value="{{.Rider.Address2}}" oninput="oid(this);" onchange="ocd(this);"><br>
    <input id="RiderTown" name="RiderTown" class="RiderTown" placeholder="town" value="{{.Rider.Town}}" oninput="oid(this);" onchange="ocd(this);"><br>
    <input id="RiderCounty" name="RiderCounty" class="RiderCounty" placeholder="county" value="{{.Rider.County}}" oninput="oid(this);" onchange="ocd(this);"><br>
	<input id="RiderPostcode" name="RiderPostcode" class="RiderPostcode" placeholder="postcode" value="{{.Rider.Postcode}}" oninput="oid(this);" onchange="ocd(this);">
    <input id="RiderCountry" name="RiderCountry" class="RiderCountry" placeholder="country" value="{{.Rider.Country}}" oninput="oid(this);" onchange="ocd(this);">
	</fieldset></div>
<fieldset class="flex field">
<div class="field">
	<label for="FreeCamping">Camping</label>
	<input type="checkbox" id="FreeCamping" name="FreeCamping" class="FreeCamping" value="FreeCamping"{{if eq .FreeCamping "Y"}} checked{{end}} onchange="oic(this);">
</div>
<div class="field">

    <label for="Route">Route</label> 
	<select id="Route" name="Route" data-chg="1" data-static="1" onchange="ocd(this);">
	    <option value="A-NCW"{{if eq .Route "A-NCW"}} selected{{end}}>North clockwise</option>
	    <option value="B-NAC"{{if eq .Route "B-NAC"}} selected{{end}}>North anticlockwise</option>
	    <option value="C-SCW"{{if eq .Route "C-SCW"}} selected{{end}}>South clockwise</option>
	    <option value="D-SAC"{{if eq .Route "D-SAC"}} selected{{end}}>South anticlockwise</option>
	    <option value="E-500CW"{{if eq .Route "E-500CW"}} selected{{end}}>500 clockwise</option>
	    <option value="F-500AC"{{if eq .Route "F-500AC"}} selected{{end}}>500 anticlockwise</option>
	</select>

</div>
<div class="field special" title="Changing status closes the form">
    
	<label for="EntrantStatus">Status</label>
	<select id="EntrantStatus" name="EntrantStatus"   data-chg="1" data-static="1" onchange="ocd(this);">
	    <option value="0"{{if eq .EntrantStatus 0}} selected{{end}}>not signed in</option>
	    <option value="2"{{if eq .EntrantStatus 2}} selected{{end}}>signed in</option>
	    <option value="4"{{if eq .EntrantStatus 4}} selected{{end}}>checked out</option>
	    <option value="8"{{if eq .EntrantStatus 8}} selected{{end}}>Finisher</option>
	    <option value="6"{{if eq .EntrantStatus 6}} selected{{end}}>DNF</option>
	    <option value="10"{{if eq .EntrantStatus 10}} selected{{end}}>Late finisher</option>
	    <option value="1"{{if eq .EntrantStatus 1}} selected{{end}}>withdrawn</option>
	</select>

</div>
<br><br>
<div class="field"><label for="Tshirt1">T-shirt 1</label> 

	<select id="Tshirt1" name="Tshirt1" class="Tshirt1"   data-chg="1" data-static="1" onchange="ocd(this);">
		<option value=""{{if eq .Tshirt1 ""}} selected{{end}}>no thanks</option>
		<option value="S"{{if eq .Tshirt1 "S"}} selected{{end}}>Small</option>
		<option value="M"{{if eq .Tshirt1 "M"}} selected{{end}}>Medium</option>
		<option value="L"{{if eq .Tshirt1 "L"}} selected{{end}}>Large</option>
		<option value="XL"{{if eq .Tshirt1 "XL"}} selected{{end}}>X-Large</option>
		<option value="XXL"{{if eq .Tshirt1 "XXL"}} selected{{end}}>XX-Large</option>
	</select>	

</div>
<div class="field"><label for="Tshirt2">T-shirt 2</label> 

	<select id="Tshirt2" name="Tshirt2" class="Tshirt2"   data-chg="1" data-static="1" onchange="ocd(this);">
		<option value=""{{if eq .Tshirt2 ""}} selected{{end}}>no thanks</option>
		<option value="S"{{if eq .Tshirt2 "S"}} selected{{end}}>Small</option>
		<option value="M"{{if eq .Tshirt2 "M"}} selected{{end}}>Medium</option>
		<option value="L"{{if eq .Tshirt2 "L"}} selected{{end}}>Large</option>
		<option value="XL"{{if eq .Tshirt2 "XL"}} selected{{end}}>X-Large</option>
		<option value="XXL"{{if eq .Tshirt2 "XXL"}} selected{{end}}>XX-Large</option>
	</select>	

</div>
<div class="field"> <label for="Patches">Patches</label> <input type="number" min="0" max="2" id="Patches" name="Patches" class="Patches" value="{{.Patches}}" oninput="oid(this);" onchange="ocd(this);"> </div>

</fieldset>
</fieldset>


<div class="tabs_area">
	<ul id="tabs">
		<li><a href="#tab_bike">Bike</a></li>
		<li><a href="#tab_nok">Emergency</a></li>
		<li><a href="#tab_money">Donations <span id="showmoney"></span></a></li>
		<li><a href="#tab_pillion">Pillion <span id="showpillion"></span></a></li>
	</ul>
</div>

<fieldset class="tabContent" id="tab_bike"><legend>Bike</legend>
<div class="field">
	<label for="Bike">Bike</label> 
	<input id="Bike" name="Bike" class="Bike" value="{{.Bike}}" oninput="oid(this);" onchange="ocd(this);">
</div>
<div class="field">
	<label for="BikeReg">Registration</label> 
	<input id="BikeReg" name="BikeReg" class="BikeReg" value="{{.BikeReg}}" oninput="oid(this);" onchange="ocd(this);">
</div>
<div class="field">
	<fieldset>
    <span class="label">Odo counts:</span>
	<input type="radio" id="OdoCountsM" name="OdoCounts" value="M"{{if ne .OdoCounts "K"}} checked{{end}} data-chg="1" data-static="1" onchange="ocd(this);"> <label for="OdoCountsM">miles</label> 
	<input type="radio" id="OdoCountsK" name="OdoCounts" value="K"{{if eq .OdoCounts "K"}} checked{{end}}  data-chg="1" data-static="1" onchange="ocd(this);"> <label for="OdoCountsK">kms</label>
	</fieldset>
</div>
<br>
<div class="field"><label for="OdoStart">Odo @ start</label> <input id="OdoStart" name="OdoStart" class="OdoStart" value="{{.OdoStart}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="OdoFinish">Odo @ finish</label> <input id="OdoFinish" name="OdoFinish" class="OdoFinish" value="{{.OdoFinish}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field" id="OdoMileage"></div>
</fieldset>
<fieldset class="tabContent" id="tab_money"><legend>Money</legend>
<div class="field">
	<label for="EntryDonation">@ entry</label> <input id="EntryDonation" name="EntryDonation" class="EntryDonation money" value="{{.FundsRaised.EntryDonation}}" oninput="moneyChg(this);"  onchange="ocd(this);">
</div>
<div class="field">
	<label for="SquiresCheque">Cheque</label> <input id="SquiresCheque" name="SquiresCheque" class="SquiresCheque money" value="{{.FundsRaised.SquiresCheque}}" oninput="moneyChg(this);" onchange="ocd(this);">
</div>
<div class="field">
	<label for="SquiresCash">Cash</label> <input id="SquiresCash" name="SquiresCash" class="SquiresCash money" value="{{.FundsRaised.SquiresCash}}" oninput="moneyChg(this);" onchange="ocd(this);">
</div>
<div class="field">
	<label for="RBLRAccount">RBLR Account</label> <input id="RBLRAccount" name="RBLRAccount" class="RBLRAccount money" value="{{.FundsRaised.RBLRAccount}}" oninput="moneyChg(this);" onchange="ocd(this);">
</div>
<div class="field">
	<label for="JustGivingAmt">Just giving</label> <input id="JustGivingAmt" name="JustGivingAmt" class="JustGivingAmt money" value="{{.FundsRaised.JustGivingAmt}}" oninput="moneyChg(this);" onchange="ocd(this);">
</div>
<div class="field">
	<label for="JustGivingURL">Just giving URL</label> <input id="JustGivingURL" name="JustGivingURL" class="JustGivingURL" value="{{.FundsRaised.JustGivingURL}}" oninput="oid(this);" onchange="ocd(this);">
</div>

</fieldset>
<fieldset class="tabContent" id="tab_nok"><legend>Emergency</legend>
<div class="field"><label for="NokName">Contact name</label> <input id="NokName" name="NokName" class="NokName" value="{{.NokName}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="NokRelation">Relationship</label> <input id="NokRelation" name="NokRelation" class="NokRelation" value="{{.NokRelation}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="NokPhone">Contact phone</label> <input id="NokPhone" name="NokPhone" class="NokPhone" value="{{.NokPhone}}" oninput="oid(this);" onchange="ocd(this);"></div>
</fieldset>
<fieldset class="tabContent" id="tab_pillion"><legend>Pillion</legend>
<div class="field"><label for="PillionLast">Last name</label> <input id="PillionLast" name="PillionLast" class="PillionLast" value="{{.Pillion.Last}}" oninput="showPillionPresent();" onchange="ocd(this);"></div>
<div class="field"><label for="PillionFirst">First name</label> <input id="PillionFirst" name="PillionFirst" class="PillionFirst" value="{{.Pillion.First}}" oninput="showPillionPresent();" onchange="ocd(this);"></div>
<div class="field"><label for="PillionIBA">IBA member</label> <input type="checkbox" id="PillionIBA" name="PillionIBA" class="PillionIBA" value="PillionIBA"{{if ne .Pillion.IBA ""}} checked{{end}} onchange="oic(this);"></div>
<div class="field"><label for="PillionRBLR">RBL Member</label> <input type="checkbox" id="PillionRBLR" name="PillionRBLR" class="PillionRBLR" value="PillionRBLR"{{if ne .Pillion.RBLR ""}} checked{{end}} onchange="oic(this);"></div>
<div class="field"><label for="PillionEmail">Email</label> <input id="PillionEmail" name="PillionEmail" class="PillionEmail" value="{{.Pillion.Email}}" oninput="oid(this);" onchange="ocd(this);"></div>
<div class="field"><label for="PillionPhone">Mobile</label> <input id="PillionPhone" name="PillionPhone" class="PillionPhone" value="{{.Pillion.Phone}}" oninput="oid(this);" onchange="ocd(this);"></div>
<br>
<div class="field"> <fieldset><legend class="small">Address</legend>
    <input id="PillionAddress1" name="PillionAddress1" class="PillionAddress1"  value="{{.Pillion.Address1}}" oninput="oid(this);" onchange="ocd(this);"><br>
    <input id="PillionAddress2" name="PillionAddress2" class="PillionAddress2"  value="{{.Pillion.Address2}}" oninput="oid(this);" onchange="ocd(this);"><br>
    <input id="PillionTown" name="PillionTown" class="PillionTown" placeholder="town" value="{{.Pillion.Town}}" oninput="oid(this);" onchange="ocd(this);"><br>
    <input id="PillionCounty" name="PillionCounty" class="PillionCounty" placeholder="county" value="{{.Pillion.County}}" oninput="oid(this);" onchange="ocd(this);"><br>
	<input id="PillionPostcode" name="PillionPostcode" class="PillionPostcode" placeholder="postcode" value="{{.Pillion.Postcode}}" oninput="oid(this);" onchange="ocd(this);">
    <input id="PillionCountry" name="PillionCountry" class="PillionCountry" placeholder="country" value="{{.Pillion.Country}}" oninput="oid(this);" onchange="ocd(this);">
	</fieldset></div>
</fieldset>



</div>
<script>` + my_tabs_js + ` setupTabs();showMoneyAmt();showPillionPresent();calcMileage();validate_entrant();setInterval(sendTransactions,1000);</script>
`

func ScanEntrant(rows *sql.Rows, e *Entrant) {

	err := rows.Scan(&e.EntrantID, &e.Rider.First, &e.Rider.Last, &e.Rider.IBA, &e.Rider.RBLR, &e.Rider.Email, &e.Rider.Phone, &e.Rider.Address1, &e.Rider.Address2, &e.Rider.Town, &e.Rider.County, &e.Rider.Postcode, &e.Rider.Country, &e.Pillion.First, &e.Pillion.Last, &e.Pillion.IBA, &e.Pillion.RBLR, &e.Pillion.Email, &e.Pillion.Phone, &e.Pillion.Address1, &e.Pillion.Address2, &e.Pillion.Town, &e.Pillion.County, &e.Pillion.Postcode, &e.Pillion.Country, &e.Bike, &e.BikeReg, &e.NokName, &e.NokRelation, &e.NokPhone, &e.OdoStart, &e.StartTime, &e.OdoFinish, &e.FinishTime, &e.EntrantStatus, &e.OdoCounts, &e.Route, &e.FundsRaised.EntryDonation, &e.FundsRaised.SquiresCash, &e.FundsRaised.SquiresCheque, &e.FundsRaised.RBLRAccount, &e.FundsRaised.JustGivingAmt, &e.Tshirt1, &e.Tshirt2, &e.Patches, &e.FreeCamping, &e.CertificateDelivered)
	checkerr(err)

}
