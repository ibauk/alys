package main

import (
	"database/sql"
	_ "embed"
	"regexp"
	"strings"
)

const JGV = "https://www.justgiving.com/page/"

//go:embed tabs.js
var my_tabs_js string

type ConfigRecord struct {
	StartTime       string
	StartCohortMins int
	ExtraCohorts    int
	RallyStatus     string
	MinDonation     int
	JustGCharities  string
	MinOdoDiff      int
	MaxOdoDiff      int
}

const ConfigSQL = `SELECT ifnull(StartTime,'05:00'),ifnull(StartCohortMins,10),ifnull(ExtraCohorts,3),ifnull(RallyStatus,'S'),ifnull(MinDonation,50),ifnull(JustGCharities,''),ifnull(MinOdoDiff,0),ifnull(MaxOdoDiff,0) FROM config`

const ConfigScreen = `
<div class="ConfigScreen">
	<div class="field">
		<label for="StartTime">Earliest start time</label> 
		<input type="time" class="StartTime" id="StartTime" name="StartTime" value="{{.StartTime}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<label for="ExtraCohorts">Number of extra cohorts</label> 
		<input type="number" min="0" max="9" class="ExtraCohorts" id="ExtraCohorts" name="ExtraCohorts" value="{{.ExtraCohorts}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<label for="StartCohortMins">Minutes between cohorts</label> 
		<input type="number" min="1" max="40" class="StartCohortMins" id="StartCohortMins" name="StartCohortMins" value="{{.StartCohortMins}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<!--
	<div class="field">
		<span class="label">State of play: </span>
		<input type="radio" id="RallyStatusS" class="RallyStatus" name="RallyStatus" value="S" {{if ne .RallyStatus "F"}} checked{{end}} data-chg="1" data-static="1" onchange="ocdcfg(this);">
		<label for="RallyStatusS">Signing-in and start before the ride</label>
		<input type="radio" id="RallyStatusF" class="RallyStatus" name="RallyStatus" value="F" {{if eq .RallyStatus "F"}} checked{{end}} data-chg="1" data-static="1" onchange="ocdcfg(this);">
		<label for="RallyStatusF">Check back in and finish after the ride</label>
	</div>
	-->
	<div class="field">
		<label for="MinDonation">Minimum donation to Poppy appeal</label>
		<input type="number" id="MinDonation" name="MinDonation" class="MinDonation" value="{{.MinDonation}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field" title="Comma-separated list of acceptable RC numbers">
		<label for="JustGCharities">Poppy appeal Reg. Charity #</label>
		<input type="text" id="JustGCharities" name="JustGCharities" class="JustGCharities" value="{{.JustGCharities}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<label for="MinOdoDiff">Minimum Odo difference</label>
		<input type="number" id="MinOdoDiff" name="MinOdoDiff" class="MinOdoDiff" value="{{.MinOdoDiff}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
	<div class="field">
		<label for="MaxOdoDiff">Maximum Odo difference</label>
		<input type="number" id="MaxOdoDiff" name="MaxOdoDiff" class="MaxOdoDiff" value="{{.MaxOdoDiff}}" oninput="oidcfg(this);" onchange="ocdcfg(this);">
	</div>
</div>
`

type Person = struct {
	First        string
	Last         string
	IBA          string
	HasIBANumber bool
	RBL          string
	Email        string
	Phone        string
	Address1     string
	Address2     string
	Town         string
	County       string
	Postcode     string
	Country      string
}

type Money = struct {
	EntryDonation string
	SquiresCheque string
	SquiresCash   string
	RBLRAccount   string
	JustGivingAmt string
	JustGivingURL string
	JustGivingPSN string
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
	CertificateAvailable string
	CertificateDelivered string
	Tshirt1              string
	Tshirt2              string
	Patches              int
	EditMode             string
	Notes                string
	Verified             string
	CertificateStatus    string
	StartTimeOnly        string
	FinishTimeOnly       string
	MinDonation          int
}

const EntrantSQL = `SELECT EntrantID,ifnull(RiderFirst,''),ifnull(RiderLast,''),ifnull(RiderIBA,''),ifnull(RiderRBL,''),ifnull(RiderEmail,''),ifnull(RiderPhone,'')
    ,ifnull(RiderAddress1,''),ifnull(RiderAddress2,''),ifnull(RiderTown,''),ifnull(RiderCounty,''),ifnull(RiderPostcode,''),ifnull(RiderCountry,'')
	,ifnull(PillionFirst,''),ifnull(PillionLast,''),ifnull(PillionIBA,''),ifnull(PillionRBL,''),ifnull(PillionEmail,''),ifnull(PillionPhone,'')
    ,ifnull(PillionAddress1,''),ifnull(PillionAddress2,''),ifnull(PillionTown,''),ifnull(PillionCounty,''),ifnull(PillionPostcode,''),ifnull(PillionCountry,'')
	,ifnull(Bike,'motorbike'),ifnull(BikeReg,'')
	,ifnull(NokName,''),ifnull(NokRelation,''),ifnull(NokPhone,'')
	,ifnull(OdoStart,''),ifnull(StartTime,''),ifnull(OdoFinish,''),ifnull(FinishTime,''),EntrantStatus,ifnull(OdoCounts,'M'),ifnull(Route,'')
	,ifnull(EntryDonation,''),ifnull(SquiresCash,''),ifnull(SquiresCheque,''),ifnull(RBLRAccount,''),ifnull(JustGivingAmt,''),ifnull(JustGivingURL,'')
	,ifnull(Tshirt1,''),ifnull(Tshirt2,''),ifnull(Patches,0),ifnull(FreeCamping,''),ifnull(CertificateDelivered,''),ifnull(CertificateAvailable,'')
	,ifnull(Notes,''),ifnull(JustGivingPSN,''),ifnull(Verified,'N'),ifnull(CertificateStatus,'N')
	 FROM entrants
`

var SigninScreenSingle = `
<div class="SigninScreenSingle">
<input type="hidden" id="EntrantID" name="EntrantID" value="{{.EntrantID}}">
<input type="hidden" id="EditMode" name="EditMode" value="{{.EditMode}}">
<input type="hidden" id="MinDonation" value="{{.MinDonation}}">

<fieldset class="tabContent" id="tab_main"><legend>&nbsp; {{.Rider.First}} {{.Rider.Last}} &nbsp;</legend>
    <div class="field special" title="Changing status closes the form">
    
	    <label for="EntrantStatus">Status</label>
	    <select id="EntrantStatus" name="EntrantStatus"   data-chg="1" data-static="1" onchange="ocd(this);" tabindex="1" autofocus>
	        <option value="0"{{if eq .EntrantStatus 0}} selected{{end}}>not signed in</option>
	        <option value="2"{{if eq .EntrantStatus 2}} selected{{end}}>signed in</option>
		    {{if eq .EditMode "signin"}}
		    {{else}}
	        <option value="4"{{if eq .EntrantStatus 4}} selected{{end}}>checked out</option>
	        <option value="8"{{if eq .EntrantStatus 8}} selected{{end}}>Finisher</option>
	        <option value="6"{{if eq .EntrantStatus 6}} selected{{end}}>DNF</option>
	        <option value="10"{{if eq .EntrantStatus 10}} selected{{end}}>Late finisher</option>
		    {{end}}
	        <option value="1"{{if eq .EntrantStatus 1}} selected{{end}}>withdrawn</option>
	    </select>

    </div>

    <div class="field" title="If switching between North & South, first untick 'Certificate available'">

        <label for="Route">Route</label> 
	    <select id="Route" name="Route" data-chg="1" data-static="1" onchange="ocd(this);" tabindex="2" >
	        <option value="A-NCW"{{if eq .Route "A-NCW"}} selected{{end}}>North clockwise</option>
	        <option value="B-NAC"{{if eq .Route "B-NAC"}} selected{{end}}>North anticlockwise</option>
	        <option value="C-SCW"{{if eq .Route "C-SCW"}} selected{{end}}>South clockwise</option>
	        <option value="D-SAC"{{if eq .Route "D-SAC"}} selected{{end}}>South anticlockwise</option>
	        <option value="E-5CW"{{if eq .Route "E-5CW"}} selected{{end}}>500 clockwise</option>
	        <option value="F-5AC"{{if eq .Route "F-5AC"}} selected{{end}}>500 anticlockwise</option>
	    </select>

    </div>




    <fieldset class="flex field">
        <div class="field">
            <label for="Tshirt1">T-shirt 1</label> 
            <select id="Tshirt1" name="Tshirt1" class="Tshirt1"   data-chg="1" data-static="1" onchange="ocd(this);" tabindex="16">
		        <option value=""{{if eq .Tshirt1 ""}} selected{{end}}>no thanks</option>
		        <option value="S"{{if eq .Tshirt1 "S"}} selected{{end}}>Small</option>
		        <option value="M"{{if eq .Tshirt1 "M"}} selected{{end}}>Medium</option>
		        <option value="L"{{if eq .Tshirt1 "L"}} selected{{end}}>Large</option>
		        <option value="XL"{{if eq .Tshirt1 "XL"}} selected{{end}}>X-Large</option>
		        <option value="XXL"{{if eq .Tshirt1 "XXL"}} selected{{end}}>XX-Large</option>
	        </select>	
        </div>

        <div class="field"><label for="Tshirt2">&nbsp; T-shirt 2</label> 
    	    <select id="Tshirt2" name="Tshirt2" class="Tshirt2"   data-chg="1" data-static="1" onchange="ocd(this);" tabindex="17">
	    	    <option value=""{{if eq .Tshirt2 ""}} selected{{end}}>no thanks</option>
		        <option value="S"{{if eq .Tshirt2 "S"}} selected{{end}}>Small</option>
		        <option value="M"{{if eq .Tshirt2 "M"}} selected{{end}}>Medium</option>
    		    <option value="L"{{if eq .Tshirt2 "L"}} selected{{end}}>Large</option>
	    	    <option value="XL"{{if eq .Tshirt2 "XL"}} selected{{end}}>X-Large</option>
		        <option value="XXL"{{if eq .Tshirt2 "XXL"}} selected{{end}}>XX-Large</option>
	        </select>	
        </div>

        <div class="field"> 
            <label for="Patches">&nbsp; Patches</label> 
            <input type="number" min="0" max="2" id="Patches" name="Patches" class="Patches" value="{{.Patches}}" oninput="oid(this);" onchange="ocd(this);" tabindex="18"> 
        </div>

        <div class="field">
			<select name="FreeCamping"  class="FreeCamping" data-chg="1" data-static="1" onchange="ocd(this);" tabindex="18">
			<option value="Y" {{if eq .FreeCamping "Y"}}selected{{end}}>Camping</option>
			<option value="N" {{if eq .FreeCamping "N"}}selected{{end}}>not camping</option>
			</select>
        </div>


		<!--
        <div class="field">
	        <label for="CertificateAvailable">Certificate available</label>
	        <input type="checkbox" id="CertificateAvailable" name="CertificateAvailable" class="CertificateAvailable" value="Y"{{if eq .CertificateAvailable "Y"}} checked{{end}} onchange="oic(this);" tabindex="19">
        </div>

        <div class="field">
	        <label for="CertificateDelivered">Certificate delivered</label>
	        <input type="checkbox" id="CertificateDelivered" name="CertificateDelivered" class="CertificateDelivered" value="Y"{{if eq .CertificateDelivered "Y"}} checked{{end}} onchange="oic(this);" tabindex="20">
        </div>
		-->

		<div class="field">
			<label for="CertificateStatus">Certificate</label>
			<select id="CertificateStatus" name="CertificateStatus" data-chg="1" data-static="1" onchange="ocd(this);" tabindex="20">
				<option value="A"{{if eq .CertificateStatus "A"}} selected{{end}}>Available</option>
				<option value="N"{{if eq .CertificateStatus "N"}} selected{{end}}>Not available</option>
				<option value="D"{{if eq .CertificateStatus "D"}} selected{{end}}>Delivered</option>
			</select>
		</div>

        {{if ge .EntrantStatus 4}}
	    <fieldset class="flex field">
		    <div class="field">
			    <label for="StartTime">Checked out @</label>
			    <input type="text" class="showtime" id="StartTime" title="{{.StartTime}}" readonly value="{{.StartTimeOnly}}" tabindex="-1">

				{{if ge .EntrantStatus 6}}
			    	{{if ge .EntrantStatus 8}}
			    	<div class="field">
						<label for="FinishTime"> Checked in @</label>
						<input type="text class="showtime" id="FinishTime" title="{{.FinishTime}}" readonly value="{{.FinishTimeOnly}}" tabindex="-1">
					</div>
			    	{{end}}
					<div class="field" title="Has this been checked by the verifier?">
						<select id="Verified" name="Verified"  data-chg="1" data-static="1" onchange="ocd(this);" tabindex="20">
						<option value="Y" {{if eq .Verified "Y"}}selected{{end}}>Verified</option>
						<option value="N" {{if eq .Verified "N"}}selected{{end}}>not verified</option>
						</select>
					</div>
				{{end}}
            </div>
	    </fieldset>
        {{end}}

    </fieldset>



</fieldset>


<div class="tabs_area">
	<ul id="tabs">
		{{if eq .EditMode "signin"}}
		<li><a tabindex="21" href="#tab_money">Donations <span id="showmoney"></span></a></li>
		<li><a tabindex="28" href="tab_notes">Notes <span id="shownotes"></span></a></li>
		{{else}}
		<li><a tabindex="21" href="tab_notes">Notes <span id="shownotes"></span></a></li>
		<li><a tabindex="28" href="#tab_money">Donations <span id="showmoney"></span></a></li>
		{{end}}
		<li><a tabindex="30" href="tab_rider" id="ridertab">Rider <span id="showrider"></span></a></li>
		<li><a tabindex="43" href="#tab_bike">Bike</a></li>
		<li><a tabindex="50" href="#tab_nok" id="noktab">Emergency</a></li>
		<li><a tabindex="54" href="#tab_pillion">Pillion <span id="showpillion"></span></a></li>
	</ul>
</div>


<fieldset class="tabContent" id="tab_money"><legend>Money</legend>
    <div class="field">
	    <label for="EntryDonation">@ entry</label> 
        <input title="Paid in via Wufoo forms on entry" id="EntryDonation" name="EntryDonation" class="EntryDonation money" value="{{.FundsRaised.EntryDonation}}" oninput="moneyChg(this);"  onchange="ocd(this);" placeholder="0.00" tabindex="22">
    </div>
    <div class="field">
	    <label for="SquiresCheque">Cheque</label> 
        <input title="Value of cheques handed in at Squires" id="SquiresCheque" name="SquiresCheque" class="SquiresCheque money" value="{{.FundsRaised.SquiresCheque}}" oninput="moneyChg(this);" onchange="ocd(this);" placeholder="0.00" tabindex="23">
    </div>
    <div class="field">
	    <label for="SquiresCash">Cash</label> 
        <input title="Value of cash handed in at Squires" id="SquiresCash" name="SquiresCash" class="SquiresCash money" value="{{.FundsRaised.SquiresCash}}" oninput="moneyChg(this);" onchange="ocd(this);" placeholder="0.00" tabindex="24">
    </div>
    <div class="field">
	    <label for="RBLRAccount">RBLR Account</label> 
        <input title="Amount paid directly (not via IBA) to RBLR account" id="RBLRAccount" name="RBLRAccount" class="RBLRAccount money" value="{{.FundsRaised.RBLRAccount}}" oninput="moneyChg(this);" onchange="ocd(this);" placeholder="0.00" tabindex="25">
    </div>
    <div class="field">
	    <label for="JustGivingAmt">JustGiving</label> 
        <input title="Amount raised using JustGiving page" id="JustGivingAmt" name="JustGivingAmt" class="JustGivingAmt money" value="{{.FundsRaised.JustGivingAmt}}" oninput="moneyChg(this);" onchange="ocd(this);" placeholder="0.00" tabindex="26" title="{{.FundsRaised.JustGivingAmt}}">
    </div>
    <div class="field" title="{{.FundsRaised.JustGivingURL}}">
	    <label for="JustGivingURL">JustGiving URL</label> 
        <input id="JustGivingURL" name="JustGivingURL" class="JustGivingURL" value="{{.FundsRaised.JustGivingURL}}" oninput="oid(this);" onchange="ocd(this);" tabindex="27" title="{{.FundsRaised.JustGivingURL}}">
    </div>
    <div class="field" title="{{.FundsRaised.JustGivingPSN}}">
	    <label for="JustGivingPSN">JustGiving PSN</label> 
        <input id="JustGivingPSN" name="JustGivingPSN" class="JustGivingPSN" value="{{.FundsRaised.JustGivingPSN}}" oninput="oid(this);" onchange="ocd(this);" tabindex="27" title="{{.FundsRaised.JustGivingPSN}}">
    </div>
</fieldset>


<fieldset class="tabContent" id="tab_notes"><legend>Notes</legend>
    <div class="field fullwidth">
	    <textarea tabindex="29" id="Notes" name="Notes" class="Notes" oninput="oid(this)" onchange="ocd(this)">{{.Notes}}</textarea>
    </div>
</fieldset>

<fieldset class="tabContent" id="tab_rider"><legend>Rider</legend>


    <div class="field">
        <label for="RiderLast">Last name</label> 
        <input id="RiderLast" name="RiderLast" class="RiderLast" value="{{.Rider.Last}}" oninput="oid(this);" onchange="ocd(this);" tabindex="31">
    </div>
    <div class="field">
        <label for="RiderFirst">First name</label> 
        <input id="RiderFirst" name="RiderFirst" class="RiderFirst" value="{{.Rider.First}}" oninput="oid(this);" onchange="ocd(this);" tabindex="32">
    </div>
    <div class="field">
        <label for="RiderEmail">Email</label> 
        <input id="RiderEmail" name="RiderEmail" class="RiderEmail" value="{{.Rider.Email}}" oninput="oid(this);" onchange="ocd(this);" tabindex="35">
    </div>
    <div class="field">
        <label for="RiderPhone">Mobile</label> 
        <input id="RiderPhone" name="RiderPhone" class="RiderPhone" value="{{.Rider.Phone}}" oninput="oid(this);" onchange="ocd(this);" tabindex="36">
		{{if ne .Rider.Phone ""}}<button onclick="callPhone('{{.Rider.Phone}}')">call</button>{{end}}
    </div>

    <div class="field">
	    <label for="RiderIBA">IBA member</label> 
	    <input type="text" id="RiderIBA" name="RiderIBA" class="RiderIBA" value="{{if .Rider.HasIBANumber}}{{.Rider.IBA}}{{else}}no{{end}}" readonly tabindex="-1">
    </div>
    <div class="field">
        <label for="RiderRBL">RBL Member</label> 
		<input type="text" id="RiderRBL" name="RiderRBL" class="RiderRBL" value="{{if eq .Rider.RBL "R"}}Rider's Branch{{else if eq .Rider.RBL "L"}}ordinary{{else}}no	{{end}}" readonly tabindex="-1">
    </div>

    <div class="field">
        <fieldset class="address"><legend class="small">Address</legend>
            <input id="RiderAddress1" name="RiderAddress1" class="RiderAddress1"  value="{{.Rider.Address1}}" oninput="oid(this);" onchange="ocd(this);" tabindex="37">
            <input id="RiderAddress2" name="RiderAddress2" class="RiderAddress2"  value="{{.Rider.Address2}}" oninput="oid(this);" onchange="ocd(this);" tabindex="38">
            <input id="RiderTown" name="RiderTown" class="RiderTown" placeholder="town" value="{{.Rider.Town}}" oninput="oid(this);" onchange="ocd(this);" tabindex="39">
            <input id="RiderCounty" name="RiderCounty" class="RiderCounty" placeholder="county" value="{{.Rider.County}}" oninput="oid(this);" onchange="ocd(this);" tabindex="40">
	        <input id="RiderPostcode" name="RiderPostcode" class="RiderPostcode" placeholder="postcode" value="{{.Rider.Postcode}}" oninput="oid(this);" onchange="ocd(this);" tabindex="41">
            <input id="RiderCountry" name="RiderCountry" class="RiderCountry" placeholder="country" value="{{.Rider.Country}}" oninput="oid(this);" onchange="ocd(this);" tabindex="42">
        </fieldset>
    </div>

</fieldset>

<fieldset class="tabContent" id="tab_bike"><legend>Bike</legend>
    <div class="field">
	    <label for="Bike">Bike</label> 
	    <input id="Bike" name="Bike" class="Bike" value="{{.Bike}}" oninput="oid(this);" onchange="ocd(this);" tabindex="44">
    </div>
    <div class="field">
	    <label for="BikeReg">Registration</label> 
	    <input id="BikeReg" name="BikeReg" class="BikeReg" value="{{.BikeReg}}" oninput="oid(this);" onchange="ocd(this);" tabindex="45">
    </div>
    <div class="field">
	    <fieldset>
           <span class="label">Odo counts:</span>
	        <input type="radio" id="OdoCountsM" name="OdoCounts" value="M"{{if ne .OdoCounts "K"}} checked{{end}} data-chg="1" data-static="1" onchange="ocd(this);" tabindex="46"> <label for="OdoCountsM">miles</label> 
	        <input type="radio" id="OdoCountsK" name="OdoCounts" value="K"{{if eq .OdoCounts "K"}} checked{{end}}  data-chg="1" data-static="1" onchange="ocd(this);" tbindex="47"> <label for="OdoCountsK">kms</label>
	    </fieldset>
    </div>


    <div class="field">
        <label for="OdoStart"> Odo @ start</label> 
        <input id="OdoStart" name="OdoStart" class="OdoStart" value="{{.OdoStart}}" oninput="oid(this);" onchange="ocd(this);" tabindex="48">
    </div>
    <div class="field">
        <label for="OdoFinish">Odo @ finish</label> 
        <input id="OdoFinish" name="OdoFinish" class="OdoFinish" value="{{.OdoFinish}}" oninput="oid(this);" onchange="ocd(this);" tabindex="49">
    </div>
    <div class="field" id="OdoMileage"></div>
</fieldset>

<fieldset class="tabContent" id="tab_nok"><legend>Emergency</legend>
    <div class="field">
        <label for="NokName">Contact name</label> 
        <input id="NokName" name="NokName" class="NokName" value="{{.NokName}}" oninput="oid(this);" onchange="ocd(this);" tabindex="51">
    </div>
    <div class="field">
        <label for="NokRelation">Relationship</label> 
        <input id="NokRelation" name="NokRelation" class="NokRelation" value="{{.NokRelation}}" oninput="oid(this);" onchange="ocd(this);" tabindex="52">
    </div>
    <div class="field">
        <label for="NokPhone">Contact phone</label> 
        <input id="NokPhone" name="NokPhone" class="NokPhone" value="{{.NokPhone}}" oninput="oid(this);" onchange="ocd(this);" tabindex="53">
		{{if ne .NokPhone ""}}<button onclick="callPhone('{{.NokPhone}}')">call</button>{{end}}
    </div>
</fieldset>

<fieldset class="tabContent" id="tab_pillion"><legend>Pillion</legend>
    <div class="field">
        <label for="PillionLast">Last name</label> 
        <input id="PillionLast" name="PillionLast" class="PillionLast" value="{{.Pillion.Last}}" oninput="showPillionPresent();" onchange="ocd(this);" tabindex="55">
    </div>
    <div class="field">
        <label for="PillionFirst">First name</label> 
        <input id="PillionFirst" name="PillionFirst" class="PillionFirst" value="{{.Pillion.First}}" oninput="showPillionPresent();" onchange="ocd(this);" tabindex="56">
    </div>
    <div class="field">
        <label for="PillionEmail">Email</label> 
        <input id="PillionEmail" name="PillionEmail" class="PillionEmail" value="{{.Pillion.Email}}" oninput="oid(this);" onchange="ocd(this);" tabindex="59">
    </div>
    <div class="field">
        <label for="PillionPhone">Mobile</label> 
        <input id="PillionPhone" name="PillionPhone" class="PillionPhone" value="{{.Pillion.Phone}}" oninput="oid(this);" onchange="ocd(this);" tabindex="60">
		{{if ne .Pillion.Phone ""}}<button onclick="callPhone('{{.Pillion.Phone}}')">call</button>{{end}}
    </div>
    <div class="field">
	    <label for="PillionIBA">IBA member</label> 
	    <input type="text" id="PillionIBA" name="PillionIBA" class="PillionIBA" value="{{if .Pillion.HasIBANumber}}{{.Pillion.IBA}}{{else}}no{{end}}" readonly tabindex="-1">
    </div>
    <div class="field">
        <label for="PillionRBL">RBL Member</label> 
		<input type="text" id="PillionRBL" name="PillionRBL" class="PillionRBL" value="{{if eq .Pillion.RBL "R"}}Rider's Branch{{else if eq .Pillion.RBL "L"}}ordinary{{else}}no	{{end}}" readonly tabindex="-1">
    </div>


    <div class="field"> 
        <fieldset class="address"><legend class="small">Address</legend>
            <input id="PillionAddress1" name="PillionAddress1" class="PillionAddress1"  value="{{.Pillion.Address1}}" oninput="oid(this);" onchange="ocd(this);" tabindex="61">
            <input id="PillionAddress2" name="PillionAddress2" class="PillionAddress2"  value="{{.Pillion.Address2}}" oninput="oid(this);" onchange="ocd(this);" tabindex="62">
            <input id="PillionTown" name="PillionTown" class="PillionTown" placeholder="town" value="{{.Pillion.Town}}" oninput="oid(this);" onchange="ocd(this);" tabindex="63">
            <input id="PillionCounty" name="PillionCounty" class="PillionCounty" placeholder="county" value="{{.Pillion.County}}" oninput="oid(this);" onchange="ocd(this);" tabindex="64">
	        <input id="PillionPostcode" name="PillionPostcode" class="PillionPostcode" placeholder="postcode" value="{{.Pillion.Postcode}}" oninput="oid(this);" onchange="ocd(this);" tabindex="65">
            <input id="PillionCountry" name="PillionCountry" class="PillionCountry" placeholder="country" value="{{.Pillion.Country}}" oninput="oid(this);" onchange="ocd(this);" tabindex="66">
        </fieldset>
	</div>
</fieldset>





<script>` + my_tabs_js + ` setupTabs();showMoneyAmt();showNotesPresent();showPillionPresent();calcMileage();validate_entrant();setInterval(sendTransactions,1000);</script>
`

func TimeOnly(dt string) string {

	dtx := strings.Split(dt, "T")
	if len(dtx) < 2 {
		return dt
	}
	return dtx[1]
}
func ScanEntrant(rows *sql.Rows, e *Entrant) {

	err := rows.Scan(&e.EntrantID, &e.Rider.First, &e.Rider.Last, &e.Rider.IBA, &e.Rider.RBL, &e.Rider.Email, &e.Rider.Phone, &e.Rider.Address1, &e.Rider.Address2, &e.Rider.Town, &e.Rider.County, &e.Rider.Postcode, &e.Rider.Country, &e.Pillion.First, &e.Pillion.Last, &e.Pillion.IBA, &e.Pillion.RBL, &e.Pillion.Email, &e.Pillion.Phone, &e.Pillion.Address1, &e.Pillion.Address2, &e.Pillion.Town, &e.Pillion.County, &e.Pillion.Postcode, &e.Pillion.Country, &e.Bike, &e.BikeReg, &e.NokName, &e.NokRelation, &e.NokPhone, &e.OdoStart, &e.StartTime, &e.OdoFinish, &e.FinishTime, &e.EntrantStatus, &e.OdoCounts, &e.Route, &e.FundsRaised.EntryDonation, &e.FundsRaised.SquiresCash, &e.FundsRaised.SquiresCheque, &e.FundsRaised.RBLRAccount, &e.FundsRaised.JustGivingAmt, &e.FundsRaised.JustGivingURL, &e.Tshirt1, &e.Tshirt2, &e.Patches, &e.FreeCamping, &e.CertificateDelivered, &e.CertificateAvailable, &e.Notes, &e.FundsRaised.JustGivingPSN, &e.Verified, &e.CertificateStatus)
	checkerr(err)

	e.Rider.HasIBANumber, _ = regexp.Match(`\d{1,6}`, []byte(e.Rider.IBA))
	e.Pillion.HasIBANumber, _ = regexp.Match(`\d{1,6}`, []byte(e.Pillion.IBA))
	e.StartTimeOnly = TimeOnly(e.StartTime)
	e.FinishTimeOnly = TimeOnly(e.FinishTime)
}
