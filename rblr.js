"use strict";

const myStackItem = "odoStack";
var timertick;



function clickTime() {
  let timeDisplay = document.querySelector('#timenow');
  console.log('Clicking time');
  clearInterval(timertick);
  if (timeDisplay.getAttribute('data-paused') != 0) {
      timeDisplay.setAttribute('data-paused',0);
      timertick = setInterval(refreshTime,timeDisplay.getAttribute('data-refresh'));
      timeDisplay.classList.remove('held');
  } else {
      timeDisplay.setAttribute('data-paused',1);
      timertick = setInterval(clickTime,timeDisplay.getAttribute('data-pause'));
      timeDisplay.classList.add('held');
  }
  console.log('Time clicked');
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
  obj.timer = setTimeout(saveOdo,3000,obj);
}

// Called during Odo capture when input is complete
function oc(obj) {
  saveOdo(obj);
}

function fix2(n) {

  if (n < 10) {
    return "0"+n;
  }
  return n;
}
function getRallyTime(dt) {

  let yy = dt.getFullYear();
  let mm = dt.getMonth() + 1;
  let dd = dt.getDate();
  let dateString = yy+'-'+fix2(mm)+'-'+fix2(dd)+"T"+dt.toLocaleTimeString('en-GB');
  return dateString.substring(0,16);
}

function parseDatetime(dt) {

  let yy = parseInt(dt.substring(0,4));
  let mm = parseInt(dt.substring(5,7)) - 1;
  let dd = parseInt(dt.substring(8,10));
  let hh = parseInt(dt.substring(11,13));
  let mn = parseInt(dt.substring(14,16));
  return new Date(yy,mm,dd,hh,mn);
}

function refreshTime() {
  sendTransactions();
  let timeDisplay = document.querySelector('#timenow');
  let dt = new Date();
  let gap = parseInt(timeDisplay.getAttribute("data-gap"));
  let xtra = parseInt(timeDisplay.getAttribute("data-xtra"));
  let newdt = getRallyTime(dt);
  let curdt = timeDisplay.getAttribute("data-time");
  console.log("Comparing "+curdt+" > "+newdt+"; xtra="+xtra+"("+gap+")");
  if (curdt >= newdt) {
    return;
  }
if (xtra > 0 && gap > 0) {
    dt = parseDatetime(curdt)
    dt = new Date(dt.getFullYear(),dt.getMonth(),dt.getDate(),dt.getHours(),dt.getMinutes()+gap);
    newdt = getRallyTime(dt);
    console.log("Choosing next slot "+newdt);
    xtra--;
    timeDisplay.setAttribute("data-xtra",xtra)
  }
  timeDisplay.setAttribute('data-time', newdt);
  let dateString = dt.toLocaleString('en-GB',{hour:"2-digit",minute:"2-digit"});
  let formattedString = dateString.replace(", ", " - ");
  timeDisplay.innerHTML = formattedString;
}




function saveOdo(obj) {

  if (obj.timer) {
    clearTimeout(obj.timer);
  }

  let timeDisplay = document.querySelector('#timenow');

  let ent = obj.getAttribute("data-e");
  let url = "putodo?e=" + ent + "&f=" + obj.name + "&v=" + obj.value + "&t="+timeDisplay.getAttribute("data-time");

  console.log(url);
  let newTrans = {};
  newTrans.url = url;
  newTrans.obj = obj.id;
  newTrans.sent = false;

  const stackx = sessionStorage.getItem(myStackItem);
  let stack = [];
  if (stackx != null) {
    stack = JSON.parse(stackx);
  }
  stack.push(newTrans);
  sessionStorage.setItem(myStackItem, JSON.stringify(stack));
  obj.classList.remove("oi");
  obj.classList.add("oc");

}
// Called periodically to send outstanding updates to backend server
function sendTransactions() {
  let stackx = sessionStorage.getItem(myStackItem);
  if (stackx == null) return;

  let stack = JSON.parse(stackx);

  let N = stack.length;

  if (N < 1) return;

  for (let i = 0; i < N; i++) {
    if (stack[i].sent) continue;

    console.log(stack[i].url);

    let xhttp = new XMLHttpRequest();
    xhttp.onerror = function () {
      return; // Probably means the network's not available
    };
    xhttp.onload = function () {
      let ok = new RegExp("\\W*ok\\W*");
      if (xhttp.status == 200) {
        console.log("{" + this.responseText.substring(0, 30) + "}");
        if (!ok.test(this.responseText)) {
          console.log("UPDATE_FAILED");
          return;
        } else {
          stack[i].sent = true;
          sessionStorage.setItem(myStackItem, JSON.stringify(stack));
          document.getElementById(stack[i].obj).classList.replace("oc", "ok");
        }
      }
    };
    console.log("url=="+stack[i].url);
    xhttp.open("GET", stack[i].url, true);
    xhttp.send();
  }
}
