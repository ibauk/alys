"use strict";

const myStackItem = "odoStack";
var timertick;

function addMoney() {
  let monies = document.getElementsByClassName("money");
  let money = 0.0;
  for (let i = 0; i < monies.length; i++) {
    let amt = parseFloat(monies[i].value);
    if (!isNaN(amt)) {
      money += amt;
    }
  }
  return money;
}

function bumpStartTime() {}

function calcMileage() {
  let mlgdiv = document.getElementById("OdoMileage");
  if (!mlgdiv) return;
  let km = document.getElementById("OdoCountsK").checked;
  let units =
    parseInt(document.getElementById("OdoFinish").value) -
    parseInt(document.getElementById("OdoStart").value);
  if (isNaN(units) || units < 1) {
    mlgdiv.innerHTML = "";
    return;
  }
  if (km) {
    const KmsPerMile = 1.60934;
    units = parseInt(units / KmsPerMile);
  }
  mlgdiv.innerHTML = " " + units + " miles";
}

function showMoneyAmt() {
  let amt = addMoney();
  let sf = document.getElementById("showmoney");
  if (sf) {
    sf.innerHTML = "£" + amt;
  }
}

function showPillionPresent() {
  let first = document.getElementById("PillionFirst");
  let last = document.getElementById("PillionLast");
  let present = first.value != "" && last.value != "";
  let ps = document.getElementById("showpillion");
  if (ps) {
    ps.innerHTML = "";
    if (present) ps.innerHTML = "&#9745;";
  }
}
function clickTime() {
  let timeDisplay = document.querySelector("#timenow");
  console.log("Clicking time");
  clearInterval(timertick);
  if (timeDisplay.getAttribute("data-paused") != 0) {
    timeDisplay.setAttribute("data-paused", 0);
    timertick = setInterval(
      refreshTime,
      timeDisplay.getAttribute("data-refresh")
    );
    timeDisplay.classList.remove("held");
  } else {
    timeDisplay.setAttribute("data-paused", 1);
    timertick = setInterval(clickTime, timeDisplay.getAttribute("data-pause"));
    timeDisplay.classList.add("held");
  }
  console.log("Time clicked");
}

function loadPage(x) {
  window.location.href = x;
}

// Called during Odo capture when input is entered
function oi(obj) {
  obj.classList.remove("oc");
  obj.classList.add("oi");

  // autosave handler
  if (obj.timer) {
    clearTimeout(obj.timer);
  }
  obj.timer = setTimeout(saveOdo, 3000, obj);
}

function oic(obj) {
  // Checkbox handler
  obj.setAttribute("data-chg", "1");
  // autosave handler
  if (obj.timer) {
    clearTimeout(obj.timer);
  }
  obj.timer = setTimeout(saveData, 1000, obj);
}

function oid(obj) {
  obj.classList.remove("oc");
  obj.classList.add("oi");
  obj.setAttribute("data-chg", "1");
  // autosave handler
  if (obj.timer) {
    clearTimeout(obj.timer);
  }
  obj.timer = setTimeout(saveData, 3000, obj);
}

function oidcfg(obj) {
  obj.classList.remove("oc");
  obj.classList.add("oi");
  obj.setAttribute("data-chg", "1");
  // autosave handler
  if (obj.timer) {
    clearTimeout(obj.timer);
  }
  obj.timer = setTimeout(saveConfig, 3000, obj);
}

// Called during Odo capture when input is complete
function oc(obj) {
  saveOdo(obj);
}

function ocd(obj) {
  if (obj.getAttribute("data-chg") == "1") {
    console.log("ocd: " + obj.name);
    saveData(obj);
  }
}

function ocdcfg(obj) {
  if (obj.getAttribute("data-chg") == "1") {
    console.log("ocd: " + obj.name);
    saveConfig(obj);
  }
}

function fix2(n) {
  if (n < 10) {
    return "0" + n;
  }
  return n;
}

function getRallyTime(dt) {
  let yy = dt.getFullYear();
  let mm = dt.getMonth() + 1;
  let dd = dt.getDate();
  let dateString =
    yy + "-" + fix2(mm) + "-" + fix2(dd) + "T" + dt.toLocaleTimeString("en-GB");
  return dateString.substring(0, 16);
}

function parseDatetime(dt) {
  let yy = parseInt(dt.substring(0, 4));
  let mm = parseInt(dt.substring(5, 7)) - 1;
  let dd = parseInt(dt.substring(8, 10));
  let hh = parseInt(dt.substring(11, 13));
  let mn = parseInt(dt.substring(14, 16));
  return new Date(yy, mm, dd, hh, mn);
}

function moneyChg(obj) {
  oic(obj);
  showMoneyAmt();
}
function refreshTime() {
  sendTransactions();
  let timeDisplay = document.querySelector("#timenow");
  let dt = new Date();
  let gap = parseInt(timeDisplay.getAttribute("data-gap"));
  let xtra = parseInt(timeDisplay.getAttribute("data-xtra"));
  let newdt = getRallyTime(dt);
  let curdt = timeDisplay.getAttribute("data-time");
  console.log(
    "Comparing " + curdt + " > " + newdt + "; xtra=" + xtra + "(" + gap + ")"
  );
  if (curdt >= newdt) {
    return;
  }
  if (xtra > 0 && gap > 0) {
    dt = parseDatetime(curdt);
    dt = new Date(
      dt.getFullYear(),
      dt.getMonth(),
      dt.getDate(),
      dt.getHours(),
      dt.getMinutes() + gap
    );
    newdt = getRallyTime(dt);
    console.log("Choosing next slot " + newdt);
    xtra--;
    timeDisplay.setAttribute("data-xtra", xtra);
  }
  timeDisplay.setAttribute("data-time", newdt);
  let dateString = dt.toLocaleString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
  });
  let formattedString = dateString.replace(", ", " - ");
  timeDisplay.innerHTML = formattedString;
}

function saveConfig(obj) {
  if (obj.getAttribute("data-static") == "") obj.setAttribute("data-chg", "");
  console.log("saveConfig: " + obj.name);
  if (obj.timer) {
    clearTimeout(obj.timer);
  }

  let val = obj.value;

  let url = encodeURI("config?" + obj.name + "=" + val);
  stackTransaction(url, obj.id);

  sendTransactions();
}

function saveFinalStatus(obj) {

  const FinisherOK = 8;

  let ent = obj.getAttribute('data-e');
  let val = obj.value;

  let url =    "putentrant?EntrantID=" + ent + "&" + obj.name + "=" + val;
  if (val!=FinisherOK) {
    url += "&CertificateAvailable=N&CertificateDelivered=N";
  }
  
  stackTransaction(encodeURI(url), obj.id);

}
function saveData(obj) {
  if (obj.getAttribute("data-static") == "") obj.setAttribute("data-chg", "");
  console.log("saveData: " + obj.name);
  if (obj.timer) {
    clearTimeout(obj.timer);
  }

  let ent = document.getElementById("EntrantID").value;
  let val = obj.value;
  switch (obj.name) {
    case "RiderPostcode":
    case "PillionPostcode":
    case "BikeReg":
    case "RiderCountry":
    case "PillionCountry":
      val = val.toUpperCase();
      break;

    case "RiderIBA":
    case "PillionIBA":
    case "RiderRBLR":
    case "PillionRBLR":
    case "FreeCamping":
    case "CertificateAvailable":
    case "CertificateDelivered":
      val = "N";
      if (obj.checked) val = "Y";
      break;

    case "EntrantStatus":
      setTimeout(endEditEntrant, 1000);
      break;

    case "OdoStart":
    case "OdoFinish":
    case "OdoKms":
    case "OdoCounts":
      calcMileage();
      break;
  }

  let url = encodeURI(
    "putentrant?EntrantID=" + ent + "&" + obj.name + "=" + val
  );
  stackTransaction(url, obj.id);

  validate_entrant();
}
function saveOdo(obj) {
  if (obj.timer) {
    clearTimeout(obj.timer);
  }

  let timeDisplay = document.querySelector("#timenow");

  let ent = obj.getAttribute("data-e");
  let url =
    "putodo?e=" +
    ent +
    "&f=" +
    obj.name +
    "&v=" +
    obj.value +
    "&t=" +
    timeDisplay.getAttribute("data-time");

  stackTransaction(url, obj.id);
}

function stackTransaction(url, objid) {
  console.log(url);
  let newTrans = {};
  newTrans.url = url;
  newTrans.objid = objid;
  newTrans.sent = false;

  const stackx = sessionStorage.getItem(myStackItem);
  let stack = [];
  if (stackx != null) {
    stack = JSON.parse(stackx);
  }
  stack.push(newTrans);
  sessionStorage.setItem(myStackItem, JSON.stringify(stack));
  /*
  obj.classList.remove("oi");
  obj.classList.add("oc");
  */
}

function sendTransactions() {
  let stackx = sessionStorage.getItem(myStackItem);
  if (stackx == null) return;

  let stack = JSON.parse(stackx);

  //console.log(stack);

  let errlog = document.getElementById('errlog')

  while (stack.length > 0) {
    let itm = stack[0];
    stack.splice(0, 1);
    sessionStorage.setItem(myStackItem, JSON.stringify(stack));
    console.log("Sending: " + itm.url);

    fetch(itm.url)
      .then((response) => {
        if (!response.ok) {
          // Handle HTTP errors
          stackTransaction(itm.url,itm.objid)
          //if (errlog){errlog.innerHTML=`HTTP error! Status: ${response.status}`}
  
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
      })
      .then((data) => {
        if (data.err) {
          // Handle JSON error field
          console.error(`Error: ${data.msg}`);
        } else {
          // Process the data if no error
          //if (errlog){errlog.innerHTML="Hello sailor: "+JSON.stringify(data)}
          console.log("Data:", data);
          document.getElementById(itm.objid).classList.replace("oi", "ok");
        }
      })
      .catch((error) => {
        // Handle network or other errors
        //if (errlog) {errlog.innerHTML="ERROR CAUGHT"}
        stackTransaction(itm.url,itm.objid)
        console.error("Fetch error:", error);
        return
      });
  }
}

function signin(e) {
  window.location = "/edit?m=signin&e=" + e;
}

function endEditEntrant() {
  let mode = document.getElementById("EditMode").value;
  switch (mode) {
    case "signin":
      window.location = "/signin";
      break;
  }
}

function validate_entrant() {
  let mustFields = [
    "RiderFirst",
    "RiderLast",
    "RiderEmail",
    "RiderPhone",
    "Bike",
    "NokName",
    "NokRelation",
    "NokPhone",
  ];

  mustFields.forEach((fld) => {
    let f = document.getElementById(fld);
    f.setAttribute("placeholder", "must not be blank");
    if (f.value == "") f.classList.add("notblank");
    else f.classList.remove("notblank");
  });
}
